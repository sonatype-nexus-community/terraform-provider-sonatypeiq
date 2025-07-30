/*
 * Copyright (c) 2019-present Sonatype, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package model

// ColorType
// --------------------------------------------
type ColorType int8

const (
	ColorDarkBlue ColorType = iota
	ColorDarkGreen
	ColorDarkPurple
	ColorDarkRed
	ColorLightBlue
	ColorLightGreen
	ColorLightPurple
	ColorLightRed
	ColorOrange
	ColorYellow
)

func (c ColorType) String() string {
	switch c {
	case ColorDarkBlue:
		return "dark-blue"
	case ColorDarkGreen:
		return "dark-green"
	case ColorDarkPurple:
		return "dark-purple"
	case ColorDarkRed:
		return "dark-red"
	case ColorLightBlue:
		return "light-blue"
	case ColorLightGreen:
		return "light-green"
	case ColorLightPurple:
		return "light-purple"
	case ColorLightRed:
		return "light-red"
	case ColorOrange:
		return "orange"
	case ColorYellow:
		return "yellow"
	}

	return "unknown"
}

func AllColors() []string {
	return []string{
		ColorDarkBlue.String(),
		ColorDarkGreen.String(),
		ColorDarkPurple.String(),
		ColorDarkRed.String(),
		ColorLightBlue.String(),
		ColorLightGreen.String(),
		ColorLightPurple.String(),
		ColorLightRed.String(),
		ColorOrange.String(),
		ColorYellow.String(),
	}
}
