package server

import (
	"context"
	"errors"
	"fmt"

	api "github.com/web-infra-dev/rslint/internal/api"
	rslintconfig "github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
)

type apiConfigModuleLoader struct {
	requester      api.Requester
	overrideConfig rslintconfig.RslintConfig
}

func (loader *apiConfigModuleLoader) LoadConfigs(ctx context.Context, request discovery.ConfigLoadBatchRequest) (discovery.ConfigLoadBatchResponse, error) {
	if loader == nil || loader.requester == nil {
		return discovery.ConfigLoadBatchResponse{}, errors.New("reverse config loading is unavailable")
	}
	msg, err := loader.requester.SendRequest(ctx, api.KindLoadConfigs, request)
	if err != nil {
		return discovery.ConfigLoadBatchResponse{}, err
	}
	var response discovery.ConfigLoadBatchResponse
	if err := msg.Decode(&response); err != nil {
		return discovery.ConfigLoadBatchResponse{}, fmt.Errorf("decode loadConfigs result: %w", err)
	}
	// API override entries are the final suffix of every loaded config. Attach
	// them at the loader boundary so discovery reachability and the published
	// catalog observe the same effective config; callers must not append them a
	// second time after discovery.
	if len(loader.overrideConfig) > 0 {
		for index := range response.Results {
			result := &response.Results[index]
			if result.Status != "loaded" {
				continue
			}
			effective := make(rslintconfig.RslintConfig, 0, len(result.Entries)+len(loader.overrideConfig))
			effective = append(effective, result.Entries...)
			effective = append(effective, loader.overrideConfig...)
			result.Entries = effective
		}
	}
	return response, nil
}

func (loader *apiConfigModuleLoader) ActivateConfigs(ctx context.Context, request discovery.ConfigActivationRequest) (discovery.ConfigActivationResponse, error) {
	if loader == nil || loader.requester == nil {
		return discovery.ConfigActivationResponse{}, errors.New("reverse config activation is unavailable")
	}
	msg, err := loader.requester.SendRequest(ctx, api.KindActivateConfigs, request)
	if err != nil {
		return discovery.ConfigActivationResponse{}, err
	}
	var response discovery.ConfigActivationResponse
	if err := msg.Decode(&response); err != nil {
		return discovery.ConfigActivationResponse{}, fmt.Errorf("decode activateConfigs result: %w", err)
	}
	return response, nil
}
