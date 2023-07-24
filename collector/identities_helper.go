// Copyright 2023 enthus GmbH
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import "time"

// convertBool2Float returns a float64 from a bool variable
func convertBool2Float(value bool) float64 {
	if value {
		return 1.0
	}

	return 0.0
}

// convertBool2Float returns a float64 from a RFC3339 string
func convertRFC33339toUnix(value string) float64 {
	t, _ := time.Parse(time.RFC3339, value)
	return float64(t.Unix())
}
