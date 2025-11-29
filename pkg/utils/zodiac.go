package utils

import (
	"time"
)

// ZodiacSign represents a zodiac sign
type ZodiacSign string

const (
	Aries       ZodiacSign = "Aries"
	Taurus      ZodiacSign = "Taurus"
	Gemini      ZodiacSign = "Gemini"
	Cancer      ZodiacSign = "Cancer"
	Leo         ZodiacSign = "Leo"
	Virgo       ZodiacSign = "Virgo"
	Libra       ZodiacSign = "Libra"
	Scorpio     ZodiacSign = "Scorpio"
	Sagittarius ZodiacSign = "Sagittarius"
	Capricorn   ZodiacSign = "Capricorn"
	Aquarius    ZodiacSign = "Aquarius"
	Pisces      ZodiacSign = "Pisces"
)

// CalculateZodiac calculates zodiac sign from date of birth
// Based on astronomical dates
func CalculateZodiac(dateOfBirth time.Time) ZodiacSign {
	month := int(dateOfBirth.Month())
	day := dateOfBirth.Day()

	switch month {
	case 1: // January
		if day <= 19 {
			return Capricorn
		}
		return Aquarius
	case 2: // February
		if day <= 18 {
			return Aquarius
		}
		return Pisces
	case 3: // March
		if day <= 20 {
			return Pisces
		}
		return Aries
	case 4: // April
		if day <= 19 {
			return Aries
		}
		return Taurus
	case 5: // May
		if day <= 20 {
			return Taurus
		}
		return Gemini
	case 6: // June
		if day <= 20 {
			return Gemini
		}
		return Cancer
	case 7: // July
		if day <= 22 {
			return Cancer
		}
		return Leo
	case 8: // August
		if day <= 22 {
			return Leo
		}
		return Virgo
	case 9: // September
		if day <= 22 {
			return Virgo
		}
		return Libra
	case 10: // October
		if day <= 22 {
			return Libra
		}
		return Scorpio
	case 11: // November
		if day <= 21 {
			return Scorpio
		}
		return Sagittarius
	case 12: // December
		if day <= 21 {
			return Sagittarius
		}
		return Capricorn
	default:
		return Aries // Default fallback
	}
}

// GetZodiacTraits returns personality traits for a zodiac sign
func GetZodiacTraits(sign ZodiacSign) string {
	traits := map[ZodiacSign]string{
		Aries:       "passionate, confident, and determined",
		Taurus:      "reliable, patient, and devoted",
		Gemini:      "adaptable, outgoing, and intelligent",
		Cancer:      "intuitive, emotional, and protective",
		Leo:         "creative, passionate, and generous",
		Virgo:       "loyal, analytical, and hardworking",
		Libra:       "diplomatic, gracious, and fair-minded",
		Scorpio:     "resourceful, brave, and passionate",
		Sagittarius: "generous, idealistic, and great sense of humor",
		Capricorn:   "responsible, disciplined, and self-controlled",
		Aquarius:    "progressive, original, and independent",
		Pisces:      "compassionate, artistic, and intuitive",
	}
	return traits[sign]
}
