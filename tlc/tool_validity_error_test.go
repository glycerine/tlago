package tlc

import "testing"

func TestToolPropertyChecksDoNotCatchJavaErrors(t *testing.T) {
	for _, check := range []struct {
		name string
		run  func(*Tool) int
	}{
		{"assumption", func(tool *Tool) int {
			tool.Assumptions = []SemanticNode{NewNumeralNode(1)}
			tool.AssumptionIsAxiom = []bool{false}
			return tool.CheckAssumptions()
		}},
		{"postcondition", func(tool *Tool) int {
			tool.PostConditionSpecs = []*Action{NewAction(NewNumeralNode(1), EmptyContext, "Post")}
			return tool.CheckPostCondition()
		}},
	} {
		for _, throws := range []bool{false, true} {
			name := check.name + "/returned"
			if throws {
				name = check.name + "/thrown"
			}
			t.Run(name, func(t *testing.T) {
				failure := NewAssertionError("property failure")
				tool := NewTool()
				tool.EvalFunc = func(*Tool, SemanticNode, ...any) (Value, error) {
					if throws {
						panic(failure)
					}
					return nil, failure
				}
				defer func() {
					if got := recover(); got != failure {
						t.Fatalf("recovered %v; want Java Error %v", got, failure)
					}
				}()
				check.run(tool)
			})
		}
	}
}

func TestToolPropertyChecksCatchThrownJavaExceptions(t *testing.T) {
	for _, check := range []struct {
		name string
		want int
		run  func(*Tool) int
	}{
		{"assumption", ECTLCAssumptionEvaluationError, func(tool *Tool) int {
			tool.Assumptions = []SemanticNode{NewNumeralNode(1)}
			tool.AssumptionIsAxiom = []bool{false}
			return tool.CheckAssumptions()
		}},
		{"postcondition", ECTLCPostconditionEvaluationError, func(tool *Tool) int {
			tool.PostConditionSpecs = []*Action{NewAction(NewNumeralNode(1), EmptyContext, "Post")}
			return tool.CheckPostCondition()
		}},
	} {
		t.Run(check.name, func(t *testing.T) {
			tool := NewTool()
			tool.EvalFunc = func(*Tool, SemanticNode, ...any) (Value, error) {
				panic(NewIllegalArgumentException("property failure"))
			}
			if got := check.run(tool); got != check.want {
				t.Fatalf("property check returned %d; want evaluation error %d", got, check.want)
			}
		})
	}
}
