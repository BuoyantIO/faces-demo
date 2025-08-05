// SPDX-FileCopyrightText: 2025 Buoyant Inc.
// SPDX-License-Identifier: Apache-2.0
//
// Copyright 2022-2025 Buoyant Inc.
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

package utils

type SmileyMap struct {
	smileys map[string]string
}

var Smileys = SmileyMap{
	smileys: map[string]string{
		"Grinning":    "&#x1F603;",
		"Sleeping":    "&#x1F634;",
		"Cursing":     "&#x1F92C;",
		"Kaboom":      "&#x1F92F;",
		"HeartEyes":   "&#x1F60D;",
		"Neutral":     "&#x1F610;",
		"RollingEyes": "&#x1F644;",
		"Screaming":   "&#x1F631;",
		"Vomiting":    "&#x1F92E;",
	},
}

// Lookup a smiley by name. If found, return the HTML entity for the smiley
// and true; if not found, return Vomiting.
func (sm *SmileyMap) Lookup(name string) string {
	if smiley, ok := sm.smileys[name]; ok {
		return smiley
	}

	// If the smiley starts with 'U+', assume it's a unicode and
	// return it as-is.

    // If the smiley starts with 'U', assume it's a unicode and return it as-is
    if len(name) > 0 && name[0] == 'U' {
        return name
    }

	// It doesn't look like a unicode and it's not in list,
	// so return Vomiting as a fallback.
	return sm.smileys["Vomiting"]
}

func (sm *SmileyMap) LookupValue(value string) string {
	for k, v := range sm.smileys {
		if v == value {
			return k
		}
	}
	// Pass through unknown values so we can see Unicodes
	return value
}

type Palette struct {
	colors map[string]string
}

// These colors are from the "Bright" color scheme shown in the "Qualitative
// Color Schemes" section of https://personal.sron.nl/~pault/. The notes about
// color pairs to avoid are from using https://davidmathlogic.com/colorblind/
// and from using the Python colorspacious module to compute distances between
// color pairs.
//
// The color names are from https://personal.sron.nl/~pault _except_ that I'm
// using "darkblue" for 4477AA, and "blue" for 66CCEE, because overall 4477AA
// turns out to cause more trouble for colorblind folks than 66CCEE.
//
// Specific problematic pairs:
//
// Protanopia/deuteranopia: darkblue/purple, green/red, grey/red, and maybe
// blue/grey (the math says it's a problem, looking at davidmathlogic seems
// like probably not?)
//
// Deuteranopia: yellow/red might be a problem, according to the math
//
// Tritanopia: this is more rare than the others, but darkblue and green are
// almost identical to this crowd, and the yellow/grey pair is troubling too.
//
// So, by default, Faces uses:
//
//    blue (66CCEE) for color workload success
//    grey (BBBBBB) for color workload error
//    purple (AA3377) for when the face workload can't talk to the color
//        workload at all
//    red (EE6677) for a color timeout and
//    yellow (CCBB44) for a latched error state
//
// and, hopefully, that's a decent compromise.

var Colors = Palette{
	colors: map[string]string{
		// Include grey/black/white because they're sometimes convenient.
		"grey":  "#BBBBBB",
		"black": "#000000",
		"white": "#FFFFFF",

		// See lots of notes above.
		"darkblue": "#4477AA",
		"blue":     "#66CCEE",
		"green":    "#228833",
		"yellow":   "#CCBB44",
		"red":      "#EE6677",
		"purple":   "#AA3377",
	},
}

func (p *Palette) Lookup(name string) string {
	if color, ok := p.colors[name]; ok {
		return color
	}

	// If the color starts with '#', assume it's a hex color code and
	// return it as-is.

	if name[0] == '#' {
		return name
	}

	// It doesn't look like a hex code and it's not a color code we know,
	// so just return yellow as a fallback.
	return p.colors["yellow"]
}

var Defaults = map[string]string{
	// Default to grey background, cursing face.
	"color":  "grey",
	"smiley": "Cursing",

	// 504 errors (GatewayTimeout) from the face workload will get handled in
	// the GUI, but from the color & smiley workloads, they should get
	// translated to a red color and a sleeping face.
	"color-504":  "red",
	"smiley-504": "Sleeping",

	// Ratelimits are yellow with an exploding head.
	"color-ratelimit":  "yellow",
	"smiley-ratelimit": "Kaboom",
}