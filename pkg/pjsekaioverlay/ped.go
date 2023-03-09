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

var WEIGHT_MAP = map[int]float64{
	0:  0,   // initialization
	1:  0,   // stage
	2:  0,   // input
	3:  1,   // tapNote
	4:  1,   // flickNote
	5:  1,   // slideStart
	6:  0.1, // slideTick
	7:  1,   // slideEnd
	8:  1,   // slideEndFlick
	9:  0,   // slideConnector
	10: 2,   // criticalTapNote
	11: 3,   // criticalFlickNote
	12: 2,   // criticalSlideStart
	13: 0.2, // criticalSlideTick
	14: 2,   // criticalSlideEnd
	15: 3,   // criticalSlideEndFlick
	16: 0,   // criticalSlideConnector
	17: 0.1, // slideHiddenTick
	18: 0.2, // traceNote
	19: 1,   // traceFlick
	20: 0.5, // criticalTraceNote
	21: 1.5, // criticalTraceFlick
	22: 1,   // traceNdFlick
	23: 0,   // judgeRenderer
	24: 0,   // longSfx
	25: 0.1, // damageNote
	26: 0.2, // traceSlideStart
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
	levelFax := float64(rating-5)*0.005 + 1

	score := 0
	entityCounter := 0
	sortedEntities := levelData.Entities
	sort.SliceStable(sortedEntities, func(i, j int) bool {
		if len(sortedEntities[i].Data.Values) == 0 || len(sortedEntities[j].Data.Values) == 0 {
			return true
		}

		return levelData.Entities[i].Data.Values[0] < levelData.Entities[j].Data.Values[0]
	})
	for _, entity := range sortedEntities {
		weight := WEIGHT_MAP[entity.Archetype]
		if weight == 0 {
			continue
		}
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
		frames = append(frames, PedFrame{
			Time:  entity.Data.Values[0],
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

		rank := "n"
		scoreX := 0.0
		if score >= 1300000 {
			rank = "s"
			scoreX = 1.0
		} else if score >= 1165000 {
			rank = "s"
			scoreX = float64((score-1165000))/(1300000-1165000)*0.110 + 0.890
		} else if score >= 940000 {
			rank = "a"
			scoreX = float64((score-940000))/(1165000-940000)*0.148 + 0.742
		} else if score >= 434000 {
			rank = "b"
			scoreX = float64((score-434000))/(940000-434000)*0.151 + 0.591
		} else if score >= 21500 {
			rank = "c"
			scoreX = float64((score-21500))/(434000-21500)*0.144 + 0.447
		} else if score >= 0 {
			rank = "n"
			scoreX = float64(score) / 21500 * 0.447
		}

		writer.Write([]byte(fmt.Sprintf("s|%f:%d:%d:%f:%s:%d\n", frame.Time, score, frameScore, scoreX, rank, i)))
	}

	return nil
}
