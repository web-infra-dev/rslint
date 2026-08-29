package unicornutil

import "testing"

func TestReplaySourceOnlyClassificationDepthBoundary(t *testing.T) {
	terminal := sourceOnlyArrayReceiverClassificationEntry{
		class: arrayClassUnknown, additionalDepth: 1, terminal: true,
	}
	classifier := arrayReceiverClassifier{recursionDepth: maxSourceOnlyArrayReceiverDepth - 1}
	if _, replay := classifier.replaySourceOnlyClassification(terminal); replay || classifier.recursionExhausted {
		t.Fatal("a terminal lower bound equal to the depth limit was replayed instead of recomputed")
	}

	classifier = arrayReceiverClassifier{recursionDepth: maxSourceOnlyArrayReceiverDepth}
	if class, replay := classifier.replaySourceOnlyClassification(terminal); !replay || class != arrayClassUnknown || !classifier.recursionExhausted ||
		classifier.recursionPeak != maxSourceOnlyArrayReceiverDepth {
		t.Fatalf("terminal overflow replay = (%v, %v, exhausted=%v, peak=%d)",
			class, replay, classifier.recursionExhausted, classifier.recursionPeak)
	}

	exact := sourceOnlyArrayReceiverClassificationEntry{
		class: arrayClassNonTarget, additionalDepth: 2,
	}
	classifier = arrayReceiverClassifier{recursionDepth: maxSourceOnlyArrayReceiverDepth - 1}
	if class, replay := classifier.replaySourceOnlyClassification(exact); !replay || class != arrayClassUnknown || !classifier.recursionExhausted ||
		classifier.recursionPeak != maxSourceOnlyArrayReceiverDepth {
		t.Fatalf("exact-cache overflow replay = (%v, %v, exhausted=%v, peak=%d)",
			class, replay, classifier.recursionExhausted, classifier.recursionPeak)
	}
}
