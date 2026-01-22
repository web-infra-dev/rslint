/**
 * RemoteTypeChecker provides a JS-side proxy to the Go-side TypeChecker
 * All method calls are forwarded via IPC to the Go backend
 */

import type { RslintServiceInterface } from './types.js';
import type {
  NodeLocation,
  NodeTypeResponse,
  NodeSymbolResponse,
  NodeSignatureResponse,
  NodeFlowNodeResponse,
  NodeInfoResponse,
} from './checker-types.js';

/**
 * RemoteTypeChecker provides access to TypeScript type information
 * through the rslint IPC interface
 */
export class RemoteTypeChecker {
  private service: RslintServiceInterface;

  constructor(service: RslintServiceInterface) {
    this.service = service;
  }

  /**
   * Get type information for a node (lazy loading)
   * @param node - NodeLocation identifying the node
   */
  async getNodeType(node: NodeLocation): Promise<NodeTypeResponse | null> {
    return this.sendRequest('checker.getNodeType', { node });
  }

  /**
   * Get symbol information for a node (lazy loading)
   * @param node - NodeLocation identifying the node
   */
  async getNodeSymbol(node: NodeLocation): Promise<NodeSymbolResponse | null> {
    return this.sendRequest('checker.getNodeSymbol', { node });
  }

  /**
   * Get signature information for a node (lazy loading)
   * @param node - NodeLocation identifying the node
   */
  async getNodeSignature(node: NodeLocation): Promise<NodeSignatureResponse | null> {
    return this.sendRequest('checker.getNodeSignature', { node });
  }

  /**
   * Get flow node information for a node (lazy loading)
   * @param node - NodeLocation identifying the node
   */
  async getNodeFlowNode(node: NodeLocation): Promise<NodeFlowNodeResponse | null> {
    return this.sendRequest('checker.getNodeFlowNode', { node });
  }

  /**
   * Get basic node information (Kind, Flags, ModifierFlags, Pos, End)
   * @param node - NodeLocation identifying the node
   */
  async getNodeInfo(node: NodeLocation): Promise<NodeInfoResponse | null> {
    return this.sendRequest('checker.getNodeInfo', { node });
  }

  /**
   * Send a request to the checker backend
   */
  private async sendRequest<T>(
    kind: string,
    data: { node: NodeLocation },
  ): Promise<T | null> {
    try {
      const response = await this.service.sendMessage(kind, data);
      return response as T;
    } catch {
      return null;
    }
  }
}
