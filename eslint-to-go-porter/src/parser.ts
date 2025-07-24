import { ClaudeResponse } from './types.js';

export class JsonStreamParser {
  private buffer = '';

  processChunk(chunk: string): ClaudeResponse[] {
    this.buffer += chunk;
    const lines = this.buffer.split('\n');
    const responses: ClaudeResponse[] = [];
    
    // Keep the last line if it's incomplete
    this.buffer = lines.pop() || '';
    
    for (const line of lines) {
      if (!line.trim()) continue;
      
      try {
        const json = JSON.parse(line) as ClaudeResponse;
        responses.push(json);
      } catch (error) {
        // Ignore lines that aren't valid JSON
        console.debug(`Skipping non-JSON line: ${line.substring(0, 50)}...`);
      }
    }
    
    return responses;
  }

  extractTextFromResponses(responses: ClaudeResponse[]): string {
    let fullText = '';
    
    for (const response of responses) {
      if (response.type === 'assistant' && response.message?.content) {
        for (const content of response.message.content) {
          if (content.text) {
            fullText += content.text;
          }
        }
      }
    }
    
    return fullText;
  }

  extractGoCode(text: string): string {
    // Look for Go code blocks in the response
    const goCodePattern = /```go\n([\s\S]*?)```/g;
    const matches = [...text.matchAll(goCodePattern)];
    
    if (matches.length > 0) {
      // Concatenate all Go code blocks
      return matches.map(match => match[1]).join('\n\n');
    }
    
    // If no code blocks found, try to extract code after certain markers
    const markerPattern = /(?:package\s+\w+[\s\S]*)/;
    const markerMatch = text.match(markerPattern);
    
    if (markerMatch) {
      return markerMatch[0];
    }
    
    throw new Error('No Go code found in response');
  }

  isErrorResponse(response: ClaudeResponse): boolean {
    return response.subtype === 'error' || 
           (response.type === 'result' && response.subtype !== 'success');
  }

  getErrorMessage(response: ClaudeResponse): string {
    return response.result?.message || 'Unknown error';
  }
}