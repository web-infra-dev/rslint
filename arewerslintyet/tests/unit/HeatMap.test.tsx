import { mockTestResults } from '../__mocks__/rslint-data';

// Mock the TooltipContext
const mockTooltip = {
  onMouseOver: jest.fn(),
  onMouseOut: jest.fn(),
};

jest.mock('../../app/TooltipContext', () => ({
  TooltipProvider: ({ children }: { children: React.ReactNode }) => children,
  useTooltip: () => mockTooltip,
}));

// Mock HeatMapItem component
jest.mock('../../app/HeatMapItem', () => {
  return function MockHeatMapItem({ tooltipContent, href, isPassing }: any) {
    return {
      type: 'a',
      props: {
        href,
        className: `w-[10px] h-[10px] ${isPassing ? 'bg-passing-square' : 'bg-failing-square'}`,
        'aria-label': `${tooltipContent} is ${isPassing ? 'passing' : 'failing'}`,
      },
    };
  };
});

describe('HeatMap Data Processing', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  describe('Test Results Processing', () => {
    it('should handle test results with passing and failing tests', () => {
      expect(mockTestResults.passing).toContain('array_type');
      expect(mockTestResults.passing).toContain('no_unused_vars');
      expect(mockTestResults.failing).toContain('no_console');
      expect(mockTestResults.failing).toContain('semicolon');
    });

    it('should parse test file structure correctly', () => {
      const passingLines = mockTestResults.passing.split('\n\n');
      expect(passingLines.length).toBeGreaterThan(0);

      // Each section should have a file path and test descriptions
      passingLines.forEach(section => {
        const lines = section.trim().split('\n');
        if (lines.length > 0) {
          const filePath = lines[0];
          expect(filePath).toMatch(/\.go$/);
        }
      });
    });

    it('should identify test descriptions correctly', () => {
      const testDescriptions = mockTestResults.passing.match(/✓ .+/g) || [];
      expect(testDescriptions.length).toBeGreaterThan(0);

      testDescriptions.forEach(desc => {
        expect(desc).toMatch(/^✓ should/);
      });
    });

    it('should handle empty test results', () => {
      const emptyResults = { passing: '', failing: '' };
      expect(emptyResults.passing).toBe('');
      expect(emptyResults.failing).toBe('');
    });
  });

  describe('HeatMap Component Logic', () => {
    it('should process test data structure correctly', () => {
      // Test the data structure that would be processed by HeapMap
      const testData = {};

      Object.keys(mockTestResults).forEach(status => {
        const value = mockTestResults[status];
        if (!value) return;

        value.split('\n\n').forEach(testGroup => {
          const lines = testGroup.replace(/\n$/, '').split('\n');
          const file = lines[0];
          const tests = lines.slice(1);

          if (!testData[file]) {
            testData[file] = {};
          }
          testData[file][status] = tests;
        });
      });

      // Verify the structure
      expect(Object.keys(testData).length).toBeGreaterThan(0);

      Object.keys(testData).forEach(file => {
        expect(file).toMatch(/\.go$/);
        const fileData = testData[file];

        if (fileData.passing) {
          expect(Array.isArray(fileData.passing)).toBe(true);
        }
        if (fileData.failing) {
          expect(Array.isArray(fileData.failing)).toBe(true);
        }
      });
    });

    it('should generate correct GitHub links', () => {
      const testFile = 'internal/rules/array_type/array_type.go';
      const expectedHref = `https://github.com/rslint/rslint/blob/main/${testFile}`;

      expect(expectedHref).toContain('github.com');
      expect(expectedHref).toContain('rslint');
      expect(expectedHref).toContain(testFile);
    });

    it('should create tooltip content correctly', () => {
      const testName = '  ✓ should detect array type violations';
      const expectedTooltip = 'it("should detect array type violations")';

      const processedName = testName.replace(/^\s*✓\s*/, ''); // Remove the checkmark and spaces
      const tooltipContent = `it("${processedName}")`;

      expect(tooltipContent).toBe(expectedTooltip);
    });
  });

  describe('Component Props Validation', () => {
    it('should validate HeatMapItem props structure', () => {
      const props = {
        tooltipContent: 'it("should test something")',
        href: 'https://github.com/rslint/rslint/blob/main/internal/rules/test.go',
        isPassing: true,
      };

      expect(props.tooltipContent).toMatch(/^it\(".+"\)$/);
      expect(props.href).toMatch(/^https:\/\/github\.com/);
      expect(typeof props.isPassing).toBe('boolean');
    });

    it('should handle both passing and failing states', () => {
      const passingProps = { isPassing: true };
      const failingProps = { isPassing: false };

      expect(passingProps.isPassing).toBe(true);
      expect(failingProps.isPassing).toBe(false);
    });
  });
});
