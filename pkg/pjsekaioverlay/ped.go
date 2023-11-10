package pjsekaioverlay

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/sevenc-nanashi/pjsekai-overlay/pkg/sonolus"
)

type PedFrame struct {
	Time  float64
	Score int
}

type BpmChange struct {
	Beat float64
	Bpm  float64
}

var WEIGHT_MAP = map[string]float64{
	"#BPM_CHANGE":    0,
	"Initialization": 0,
	"InputManager":   0,
	"Stage":          0,

	"NormalTapNote":   1,
	"CriticalTapNote": 3,

	"NormalFlickNote":   1,
	"CriticalFlickNote": 3,

	"NormalSlideStartNote":   1,
	"CriticalSlideStartNote": 2,

	"NormalSlideEndNote":   1,
	"CriticalSlideEndNote": 2,

	"NormalSlideEndFlickNote":   1,
	"CriticalSlideEndFlickNote": 3,

	"HiddenSlideTickNote":   0,
	"NormalSlideTickNote":   0.1,
	"CriticalSlideTickNote": 0.1,

	"IgnoredSlideTickNote":          0.1,
	"NormalAttachedSlideTickNote":   0.1,
	"CriticalAttachedSlideTickNote": 0.1,

	"NormalSlideConnector":   0,
	"CriticalSlideConnector": 0,

	"SimLine": 0,

	"NormalSlotEffect":       0,
	"SlideSlotEffect":        0,
	"FlickSlotEffect":        0,
	"CriticalSlotEffect":     0,
	"NormalSlotGlowEffect":   0,
	"SlideSlotGlowEffect":    0,
	"FlickSlotGlowEffect":    0,
	"CriticalSlotGlowEffect": 0,

	"NormalTraceNote":   0.1,
	"CriticalTraceNote": 0.1,

	"NormalTraceSlotEffect":     0,
	"NormalTraceSlotGlowEffect": 0,

	"DamageNote":           0.1,
	"DamageSlotEffect":     0,
	"DamageSlotGlowEffect": 0,

	"NormalTraceFlickNote":         0.5,
	"CriticalTraceFlickNote":       0.5,
	"NonDirectionalTraceFlickNote": 0.5,

	"NormalTraceSlideStartNote":   0.1,
	"NormalTraceSlideEndNote":     0.1,
	"CriticalTraceSlideStartNote": 0.1,
	"CriticalTraceSlideEndNote":   0.1,

	"TimeScaleGroup":  0,
	"TimeScaleChange": 0,
}

func getValueFromData(data []sonolus.LevelDataEntityValue, name string) (float64, error) {
	for _, value := range data {
		if value.Name == name {
			return value.Value, nil
		}
	}
	return 0, fmt.Errorf("value not found: %s", name)
}

func getTimeFromBpmChanges(bpmChanges []BpmChange, beat float64) float64 {
	ret := 0.0
	for i, bpmChange := range bpmChanges {
		if i == len(bpmChanges)-1 {
			ret += (beat - bpmChange.Beat) * (60 / bpmChange.Bpm)
			break
		}
		nextBpmChange := bpmChanges[i+1]
		if beat >= bpmChange.Beat && beat < nextBpmChange.Beat {
			ret += (beat - bpmChange.Beat) * (60 / bpmChange.Bpm)
			break
		} else if beat >= nextBpmChange.Beat {
			ret += (nextBpmChange.Beat - bpmChange.Beat) * (60 / bpmChange.Bpm)
		} else {
			break
		}
	}
	return ret
}

func CalculateScore(levelInfo sonolus.LevelInfo, levelData sonolus.LevelData, power int) []PedFrame {
	rating := levelInfo.Rating
	framesLen := 0
	var weightedNotesCount float64 = 0
	for _, entity := range levelData.Entities {
		weight := WEIGHT_MAP[entity.Archetype]
		if weight == 0 {
			continue
		}
		weightedNotesCount += weight
		framesLen += 1
	}

	frames := make([]PedFrame, 0, int(weightedNotesCount)+1)
	frames = append(frames, PedFrame{Time: 0, Score: 0})
	bpmChanges := ([]BpmChange{})
	levelFax := float64(rating-5)*0.005 + 1

	score := 0
	entityCounter := 0
	noteEntities := ([]sonolus.LevelDataEntity{})

	for _, entity := range levelData.Entities {
		weight := WEIGHT_MAP[entity.Archetype]
		if weight > 0.0 && len(entity.Data) > 0 {
			noteEntities = append(noteEntities, entity)
		} else if entity.Archetype == "#BPM_CHANGE" {
			beat, err := getValueFromData(entity.Data, "#BEAT")
			if err != nil {
				continue
			}
			bpm, err := getValueFromData(entity.Data, "#BPM")
			if err != nil {
				continue
			}
			bpmChanges = append(bpmChanges, BpmChange{
				Beat: beat,
				Bpm:  bpm,
			})
		}
	}
	sort.SliceStable(noteEntities, func(i, j int) bool {
		return noteEntities[i].Data[0].Value < noteEntities[j].Data[0].Value
	})
	sort.SliceStable(bpmChanges, func(i, j int) bool {
		return bpmChanges[i].Beat < bpmChanges[j].Beat
	})
	for _, entity := range noteEntities {
		weight := WEIGHT_MAP[entity.Archetype]
		entityCounter += 1

		score += int(
			(float64(power) / weightedNotesCount) * // Team power / weighted notes count
				4 * // Constant
				weight * // Note weight
				1 * // Judge weight (Always 1)
				levelFax * // Level fax
				(float64(entityCounter/100)/100 + 1) * // Combo fax
				1, // Skill fax (Always 1)
		)
		beat, err := getValueFromData(entity.Data, "#BEAT")
		if err != nil {
			continue
		}
		frames = append(frames, PedFrame{
			Time:  getTimeFromBpmChanges(bpmChanges, beat) + levelData.BgmOffset,
			Score: score,
		})
	}

	return frames
}

func WritePedFile(frames []PedFrame, assets string, ap bool, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイルの作成に失敗しました（%s）", err)
	}
	defer file.Close()

	writer := io.Writer(file)

	writer.Write([]byte(fmt.Sprintf("p|%s\n", assets)))
	writer.Write([]byte(fmt.Sprintf("a|%s\n", strconv.FormatBool(ap))))
	writer.Write([]byte(fmt.Sprintf("v|%s\n", Version)))
	writer.Write([]byte(fmt.Sprintf("u|%d\n", time.Now().Unix())))

	lastScore := 0
	for i, frame := range frames {
		score := frame.Score
		frameScore := score - lastScore
		lastScore = frame.Score

		// 161, 215, 267, 320, 357

		rank := "n"
		scoreX := 0.0
		if score >= 1300000 {
			rank = "s"
			scoreX = 357
		} else if score >= 1165000 {
			rank = "s"
			scoreX = float64((score-1165000))/(1300000-1165000)*37 + 320
		} else if score >= 940000 {
			rank = "a"
			scoreX = float64((score-940000))/(1165000-940000)*53 + 267
		} else if score >= 434000 {
			rank = "b"
			scoreX = float64((score-434000))/(940000-434000)*53 + 215
		} else if score >= 21500 {
			rank = "c"
			scoreX = float64((score-21500))/(434000-21500)*54 + 161
		} else if score >= 0 {
			rank = "d"
			scoreX = float64(score) / 21500 * 160
		}

		writer.Write([]byte(fmt.Sprintf("s|%f:%d:%d:%f:%s:%d\n", frame.Time, score, frameScore, scoreX/357, rank, i)))
	}

	return nil
}
