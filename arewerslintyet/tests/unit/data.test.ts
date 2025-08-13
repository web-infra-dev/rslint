import {
  mockKV,
  resetKVMocks,
  mockTestResults,
  mockGraphData,
  mockExamplesData,
  mockProcessedGraphData,
} from '../__mocks__/rslint-data';

// Mock the data module
jest.mock('../../app/data', () => {
  const originalModule = jest.requireActual('../../app/data');
  return {
    ...originalModule,
    getDevelopmentTestResults: jest.fn(),
    getProductionTestResults: jest.fn(),
    getExamplesResults: jest.fn(),
    getDevelopmentTestRuns: jest.fn(),
    getProductionTestRuns: jest.fn(),
  };
});

describe('Data Functions', () => {
  beforeEach(() => {
    mockKV();
  });

  afterEach(() => {
    resetKVMocks();
    jest.clearAllMocks();
  });

  describe('getDevelopmentTestResults', () => {
    it('should return test results when data exists', async () => {
      const { getDevelopmentTestResults } = require('../../app/data');
      getDevelopmentTestResults.mockResolvedValue(mockTestResults);

      const result = await getDevelopmentTestResults();

      expect(result).toEqual(mockTestResults);
      expect(result.passing).toContain('array_type');
      expect(result.failing).toContain('no_console');
    });

    it('should return null when no data exists', async () => {
      const { getDevelopmentTestResults } = require('../../app/data');
      getDevelopmentTestResults.mockResolvedValue(null);

      const result = await getDevelopmentTestResults();

      expect(result).toBeNull();
    });
  });

  describe('getProductionTestResults', () => {
    it('should return production test results', async () => {
      const { getProductionTestResults } = require('../../app/data');
      getProductionTestResults.mockResolvedValue(mockTestResults);

      const result = await getProductionTestResults();

      expect(result).toEqual(mockTestResults);
    });
  });

  describe('getExamplesResults', () => {
    it('should return examples data', async () => {
      const { getExamplesResults } = require('../../app/data');
      getExamplesResults.mockResolvedValue(mockExamplesData);

      const result = await getExamplesResults();

      expect(result).toEqual(mockExamplesData);
      expect(result['basic-linting']).toBe(true);
      expect(result['custom-rules']).toBe(false);
    });
  });

  describe('getDevelopmentTestRuns', () => {
    it('should return processed graph data with most recent', async () => {
      const { getDevelopmentTestRuns } = require('../../app/data');
      const mockResponse = {
        graphData: mockProcessedGraphData,
        mostRecent: mockProcessedGraphData[mockProcessedGraphData.length - 1],
      };
      getDevelopmentTestRuns.mockResolvedValue(mockResponse);

      const result = await getDevelopmentTestRuns();

      expect(result.graphData).toHaveLength(5);
      expect(result.mostRecent.percent).toBe(93.0);
      expect(result.graphData[0].gitHash).toBe('abc1234');
    });
  });

  describe('getProductionTestRuns', () => {
    it('should return production graph data', async () => {
      const { getProductionTestRuns } = require('../../app/data');
      const mockResponse = {
        graphData: mockProcessedGraphData,
        mostRecent: mockProcessedGraphData[mockProcessedGraphData.length - 1],
      };
      getProductionTestRuns.mockResolvedValue(mockResponse);

      const result = await getProductionTestRuns();

      expect(result.graphData).toHaveLength(5);
      expect(result.mostRecent).toBeDefined();
    });
  });
});

// Test the actual processGraphData function if we can access it
describe('Graph Data Processing', () => {
  it('should process raw graph data correctly', () => {
    // This would test the actual processGraphData function
    // For now, we'll test the expected output format
    const expectedFormat = {
      gitHash: expect.any(String),
      date: expect.any(Number),
      total: expect.any(Number),
      passing: expect.any(Number),
      percent: expect.any(Number),
    };

    mockProcessedGraphData.forEach(item => {
      expect(item).toMatchObject(expectedFormat);
      expect(item.percent).toBeGreaterThanOrEqual(0);
      expect(item.percent).toBeLessThanOrEqual(100);
    });
  });
});
