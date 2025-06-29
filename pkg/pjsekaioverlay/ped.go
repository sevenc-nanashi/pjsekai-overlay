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
	Score float64
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
	"CriticalTapNote": 2,

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
	"CriticalSlideTickNote": 0.2,

	"IgnoredSlideTickNote":          0.1,
	"NormalAttachedSlideTickNote":   0.1,
	"CriticalAttachedSlideTickNote": 0.2,

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
	"CriticalTraceNote": 0.2,

	"NormalTraceSlotEffect":     0,
	"NormalTraceSlotGlowEffect": 0,

	"DamageNote":           0.1,
	"DamageSlotEffect":     0,
	"DamageSlotGlowEffect": 0,

	"NormalTraceFlickNote":         1,
	"CriticalTraceFlickNote":       3,
	"NonDirectionalTraceFlickNote": 1,

	"NormalTraceSlideStartNote":   0.1,
	"NormalTraceSlideEndNote":     0.1,
	"CriticalTraceSlideStartNote": 0.2,
	"CriticalTraceSlideEndNote":   0.2,

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
	var weightedNotesCount float64 = 0
	for _, entity := range levelData.Entities {
		weight := WEIGHT_MAP[entity.Archetype]
		if weight == 0 {
			continue
		}
		weightedNotesCount += weight
	}

	frames := make([]PedFrame, 0, int(weightedNotesCount)+1)
	frames = append(frames, PedFrame{Time: 0, Score: 0})
	bpmChanges := ([]BpmChange{})
	levelFax := float64(rating-5)*0.005 + 1
	comboFax := 1.0

	score := 0.0
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
		if entityCounter%100 == 1 && entityCounter > 1 {
			comboFax += 0.01
		}
		if comboFax > 1.1 {
			comboFax = 1.1
		}

		score += ((float64(power) / weightedNotesCount) * // Team power / weighted notes count
				4 * // Constant
				weight * // Note weight
				1 * // Judge weight (Always 1)
				levelFax * // Level fax
				comboFax * // Combo fax
				1) // Skill fax (Always 1)
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

func WritePedFile(frames []PedFrame, assets string, ap bool, path string, levelInfo sonolus.LevelInfo) error {
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

	lastScore := 0.0
	rating := levelInfo.Rating
	for i, frame := range frames {
		score := frame.Score
		frameScore := score - lastScore
		lastScore = frame.Score

		// 161, 215, 267, 320, 357

		rank := "n"
		scoreX := 0.0

		// rank
		if rating < 5 {
			rating = 5
		} else if rating > 40 {
			rating = 40
		}

		rankBorder := float64(1200000 + (rating-5)*4100)
		rankS := float64(1040000 + (rating-5)*5200)
		rankA := float64(840000 + (rating-5)*4200)
		rankB := float64(400000 + (rating-5)*2000)
		rankC := float64(20000 + (rating-5)*100)

		// bar
		if score >= rankBorder {
			rank = "s"
			scoreX = 357
		} else if score >= rankS {
			rank = "s"
			scoreX = (float64((score-rankS))/float64((rankBorder-rankS)))*37 + 320
		} else if score >= rankA {
			rank = "a"
			scoreX = (float64((score-rankA))/float64((rankS-rankA)))*53 + 267
		} else if score >= rankB {
			rank = "b"
			scoreX = (float64((score-rankB))/float64((rankA-rankB)))*53 + 215
		} else if score >= rankC {
			rank = "c"
			scoreX = (float64((score-rankC))/float64((rankB-rankC)))*54 + 161
		} else {
			rank = "d"
			scoreX = (float64(score) / float64(rankC)) * 160
		}
		
		time := frame.Time
		if time == 0 && i > 0 {
			time = frames[i-1].Time + 0.000001
		}
		
		writer.Write([]byte(fmt.Sprintf("s|%f:%f:%f:%f:%s:%d\n", time, score, frameScore, scoreX/357, rank, i)))
	}

	return nil
}
