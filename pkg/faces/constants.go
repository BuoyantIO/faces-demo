// SPDX-FileCopyrightText: 2024 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2024 Buoyant Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.  You may obtain
// a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package faces

var Smileys = map[string]string{
	"Smiling":     "&#x1F603;",
	"Sleeping":    "&#x1F634;",
	"Cursing":     "&#x1F92C;",
	"Kaboom":      "&#x1F92F;",
	"HeartEyes":   "&#x1F60D;",
	"Neutral":     "&#x1F610;",
	"RollingEyes": "&#x1F644;",
	"Screaming":   "&#x1F631;",
}

var Defaults = map[string]string{
	// Default to grey background, cursing face.
	"color":  "grey",
	"smiley": Smileys["Cursing"],

	// 504 errors (GatewayTimeout) from the face workload will get handled in
	// the GUI, but from the color & smiley workloads, they should get
	// translated to a pink color or a sleeping face.
	"color-504":  "pink",
	"smiley-504": Smileys["Sleeping"],

	// Ratelimits are pink with an exploding head.
	"color-ratelimit":  "pink",
	"smiley-ratelimit": Smileys["Kaboom"],
}
