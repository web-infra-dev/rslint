package ipc

import (
	"context"
	"time"
)

// Setup is shared by the session, not owned by whichever request first needs
// bytes. Bound it independently of application deadlines (which may be absent).
const memorySetupTimeout = 5 * time.Second

type memoryPrepare struct {
	Configuration MemoryConfiguration `json:"configuration"`
	SocketPath    string              `json:"socketPath,omitempty"`
}

func (c *Channel) ensureMemory(ctx context.Context, attachments []Attachment) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	eligible := false
	for _, value := range attachments {
		if size := value.length(); size > 0 && size <= MemorySlotCount*MemorySlotSize {
			eligible = true
			break
		}
	}
	if !eligible {
		return nil
	}
	c.mu.Lock()
	if c.closed {
		err := c.closeErr
		c.mu.Unlock()
		return err
	}
	if !c.memoryCapable {
		c.mu.Unlock()
		return nil
	}
	if c.memorySetupDone == nil {
		c.memorySetupDone = make(chan struct{})
		go c.setupMemory()
	}
	done := c.memorySetupDone
	c.mu.Unlock()
	select {
	case <-done:
		c.mu.Lock()
		err := c.closeErr
		c.mu.Unlock()
		return err
	case <-ctx.Done():
		return ctx.Err()
	case <-c.done:
		return c.closeErr
	}
}

func (c *Channel) setupMemory() {
	defer close(c.memorySetupDone)
	ctx, cancel := context.WithTimeout(c.inCtx, memorySetupTimeout)
	defer cancel()
	bootstrap, err := listenMemoryBootstrap(ctx)
	if err != nil {
		return // The session stays inline after an unavailable platform or socket.
	}
	defer bootstrap.close()
	committed := false
	defer func() {
		if !committed && ctx.Err() != context.Canceled {
			// Ordered after prepare, including a late reply after setup timed out.
			// A transport write failure remains terminal via SendNotification.
			_ = c.SendNotification(KindTransportAbort, nil)
		}
	}()
	response, err := c.SendRequest(ctx, KindTransportPrepare, memoryPrepare{
		Configuration: memoryConfiguration(), SocketPath: bootstrap.path(),
	})
	if err != nil {
		return
	}
	var descriptor MemoryMapping
	if err := response.Decode(&descriptor); err != nil || descriptor.Version != MemoryVersion {
		return
	}
	mapping, err := bootstrap.receive(ctx, descriptor)
	if err != nil {
		return
	}
	memory := &memoryPool{mapping: mapping}
	defer func() {
		if !committed {
			_ = memory.close()
		}
	}()
	if ctx.Err() != nil {
		return
	}
	// No request can publish ranges until the commit frame precedes it on the
	// normal stream. Neither the read loop nor unrelated requests wait on setup.
	if err := c.SendNotification(KindTransportCommit, nil); err != nil {
		return
	}
	c.mu.Lock()
	if !c.closed {
		c.memory = memory
		committed = true
	}
	c.mu.Unlock()
}
