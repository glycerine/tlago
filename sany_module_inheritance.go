// Copyright (c) 2003 Compaq Corporation. All rights reserved.
// Portions Copyright (c) 2003 Microsoft Corporation. All rights reserved.
package tlago

import "github.com/glycerine/tlago/tlc"

func (module *sanySemModuleNode) copyAssumes(extendee *sanySemModuleNode) {
	if module == nil || extendee == nil {
		panic(tlc.NewNullPointerException())
	}
	for i := 0; i < len(extendee.assumptionVec); i++ {
		module.assumptionVec = append(module.assumptionVec, extendee.assumptionVec[i])
	}
}
func (module *sanySemModuleNode) copyTheorems(extendee *sanySemModuleNode) {
	if module == nil || extendee == nil {
		panic(tlc.NewNullPointerException())
	}
	for i := 0; i < len(extendee.theoremVec); i++ {
		module.theoremVec = append(module.theoremVec, extendee.theoremVec[i])
	}
}
func (module *sanySemModuleNode) copyTopLevel(extendee *sanySemModuleNode) {
	if module == nil || extendee == nil {
		panic(tlc.NewNullPointerException())
	}
	for i := 0; i < len(extendee.topLevelVec); i++ {
		module.topLevelVec = append(module.topLevelVec, extendee.topLevelVec[i])
	}
}
func (module *sanySemModuleNode) createExtendeeArray(vector []*sanySemModuleNode) {
	if module == nil || vector == nil {
		panic(tlc.NewNullPointerException())
	}
	module.extendees = make([]*sanySemModuleNode, len(vector))
	copy(module.extendees, vector)
}
func (module *sanySemModuleNode) getExtendedModuleSet(recursively bool) map[*sanySemModuleNode]struct{} {
	if module == nil {
		panic(tlc.NewNullPointerException())
	}
	if result := module.depthAllExtendees[recursively]; result != nil {
		return result
	}
	if module.extendees == nil {
		panic(tlc.NewNullPointerException())
	}
	result := make(map[*sanySemModuleNode]struct{})
	for _, extendee := range module.extendees {
		result[extendee] = struct{}{}
		if recursively {
			if extendee == nil {
				panic(tlc.NewNullPointerException())
			}
			for inherited := range extendee.getExtendedModuleSet(true) {
				result[inherited] = struct{}{}
			}
		}
	}
	module.depthAllExtendees[recursively] = result
	return result
}
func (module *sanySemModuleNode) extendsModule(other *sanySemModuleNode) bool {
	_, present := module.getExtendedModuleSet(true)[other]
	return present
}
