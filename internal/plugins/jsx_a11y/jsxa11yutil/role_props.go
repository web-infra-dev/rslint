package jsxa11yutil

import "sort"

// The data behind RoleAriaProps lives in role_props_generated.go, produced by
// scripts/gen-aria-role-props.mjs. It mirrors aria-query `rolesMap[role].props`
// keys for every ARIA role (including abstract, DPUB and Graphics roles). The
// set of supported props for a role is the union of the role itself and all
// its superRoles in aria-query — i.e. the inheritance chain has already been
// flattened in the source data.
//
// Used by jsx-a11y/role-supports-aria-props to compute the set of ARIA props
// that are NOT supported by a given role:
//
//	const invalidAriaPropsForRole = new Set(aria.keys().filter(
//	  (attr) => !(attr in propKeyValues)));
//
// Abstract roles (`command`, `composite`, `input`, `landmark`, `range`,
// `roletype`, `section`, `sectionhead`, `select`, `structure`, `widget`,
// `window`) are intentionally INCLUDED — upstream `roles.get(roleValue)`
// matches them when the user writes `role="command"` etc., and the rule
// validates the supported-props set against them like any other role.
//
// Source: https://github.com/A11yance/aria-query/tree/v5.3.2/src/etc/roles

// RoleAriaProps is the set of ARIA properties a role supports, one bit per
// property in ariaRolePropNames.
type RoleAriaProps uint64

// LookupRoleAriaProps returns the supported-property set of an ARIA role. The
// lookup is case-sensitive. ok is false for a role aria-query does not know,
// which is distinct from a known role whose set is empty.
func LookupRoleAriaProps(role string) (props RoleAriaProps, ok bool) {
	i := sort.SearchStrings(ariaRoleNames[:], role)
	if i == len(ariaRoleNames) || ariaRoleNames[i] != role {
		return 0, false
	}
	return RoleAriaProps(ariaRoleMasks[i]), true
}

// Supports reports whether the role supports the named ARIA property. The
// name is matched case-sensitively; a name that is not an ARIA property is
// never supported.
func (p RoleAriaProps) Supports(prop string) bool {
	i := sort.SearchStrings(ariaRolePropNames[:], prop)
	return i < len(ariaRolePropNames) && ariaRolePropNames[i] == prop && p&(1<<uint(i)) != 0
}

// Names returns the supported property names in sorted order.
func (p RoleAriaProps) Names() []string {
	var names []string
	for i, name := range ariaRolePropNames {
		if p&(1<<uint(i)) != 0 {
			names = append(names, name)
		}
	}
	return names
}
