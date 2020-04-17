package persistence

func revertIntoMap(slice []interface{}) map[interface{}]struct{} {
	result := make(map[interface{}]struct{}, len(slice))
	for _, e := range slice {
		result[e] = struct{}{}
	}
	return result
}

func mapKeyDiff(m1, m2 map[interface{}]struct{}) (result []interface{}) {
	for k := range m1 {
		if _, exists := m2[k]; !exists {
			result = append(result, k)
		}
	}
	return
}

type equaler interface {
	equal(e equaler) bool
}

func mapEquals(m1, m2 map[string]interface{}) bool {
	if len(m1) != len(m2) {
		return false
	}

	for k, vm1 := range m1 {
		if vm2, exists := m2[k]; !exists {
			return false
		} else {
			vm1eq, vm1IsEqualer := vm1.(equaler)
			vm2eq, vm2IsEqualer := vm2.(equaler)

			if vm1IsEqualer != vm2IsEqualer {
				return false
			}

			if vm1IsEqualer && vm2IsEqualer {
				if !vm1eq.equal(vm2eq) {
					return false
				}
			} else {
				if vm1 != vm2 {
					return false
				}
			}
		}
	}
	return true
}
