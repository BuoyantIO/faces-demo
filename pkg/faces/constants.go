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

// colorblind-friendly colors from the Tol palette
var Colors = map[string]string{
	"grey":    "grey",
	"purple":  "rgb(48 34 130)",
	"green":   "rgb(55 117 59)",
	"cyan":    "rgb(55 117 59)",   // too similar to pink and grey, avoid
	"blue":    "rgb(151 202 234)", // light blue
	"yellow":  "rgb(218 204 130)",
	"pink":    "rgb(191 108 120)",
	"hotpink": "rgb(158 75 149)",
	"red":     "rgb(125 42 83)", // dark magenta
}

var Defaults = map[string]string{
	// Default to grey background, cursing face.
	"color":  Colors["grey"],
	"smiley": Smileys["Cursing"],

	// 504 errors (GatewayTimeout) from the face workload will get handled in
	// the GUI, but from the color & smiley workloads, they should get
	// translated to a pink color and a sleeping face.
	"color-504":  Colors["pink"],
	"smiley-504": Smileys["Sleeping"],

	// Ratelimits are pink with an exploding head.
	"color-ratelimit":  Colors["pink"],
	"smiley-ratelimit": Smileys["Kaboom"],
}
