package telegram_bot

import (
	"fmt"
	"strings"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

// поведенческие сигналы интервью
type AntiCheatReport struct {
	AvgResponseTimeSec  float64  `json:"avg_response_time_sec"`
	MinResponseTimeSec  int      `json:"min_response_time_sec"`
	MaxResponseTimeSec  int      `json:"max_response_time_sec"`
	FastResponseCount   int      `json:"fast_response_count"`   // < 5 сек
	SlowResponseCount   int      `json:"slow_response_count"`
	TotalMessages       int      `json:"total_messages"`
	VoiceMessageCount   int      `json:"voice_message_count"`
	TextMessageCount    int      `json:"text_message_count"`
	VoiceToTextRatio    float64  `json:"voice_to_text_ratio"`
	Flags               []string `json:"flags"`
}

// анализ диалога на подозрительные паттерны
func collectAntiCheatSignals(conversation []models.ConversationMessage) *AntiCheatReport {
	report := &AntiCheatReport{
		MinResponseTimeSec: 9999,
	}

	var totalResponseTime int
	var responseCount int

	for _, msg := range conversation {
		if msg.Role != "candidate" {
			continue
		}

		report.TotalMessages++

		if msg.MessageType == "voice" {
			report.VoiceMessageCount++
		} else {
			report.TextMessageCount++
		}

		if msg.ResponseTimeSec > 0 {
			responseCount++
			totalResponseTime += msg.ResponseTimeSec

			if msg.ResponseTimeSec < report.MinResponseTimeSec {
				report.MinResponseTimeSec = msg.ResponseTimeSec
			}
			if msg.ResponseTimeSec > report.MaxResponseTimeSec {
				report.MaxResponseTimeSec = msg.ResponseTimeSec
			}

			// подозрительно быстрый ответ
			if msg.ResponseTimeSec < 5 && len(msg.Content) > 50 {
				report.FastResponseCount++
			}

			// медленные ответы не флагуем — кандидат может думать
		}
	}

	if responseCount > 0 {
		report.AvgResponseTimeSec = float64(totalResponseTime) / float64(responseCount)
	}
	if report.MinResponseTimeSec == 9999 {
		report.MinResponseTimeSec = 0
	}

	if report.TotalMessages > 0 {
		report.VoiceToTextRatio = float64(report.VoiceMessageCount) / float64(report.TotalMessages)
	}

	// формируем флаги
	if report.FastResponseCount >= 3 {
		report.Flags = append(report.Flags, "multiple_fast_responses")
	}

	// резкая смена стиля
	if detectStyleShift(conversation) {
		report.Flags = append(report.Flags, "style_shift_detected")
	}

	if report.Flags == nil {
		report.Flags = []string{}
	}

	return report
}

// резкое изменение стиля письма в ходе диалога
func detectStyleShift(conversation []models.ConversationMessage) bool {
	var candidateMsgs []string
	for _, msg := range conversation {
		if msg.Role == "candidate" && len(msg.Content) > 30 {
			candidateMsgs = append(candidateMsgs, msg.Content)
		}
	}

	if len(candidateMsgs) < 4 {
		return false
	}

	// средняя длина слова: первая vs вторая половина
	mid := len(candidateMsgs) / 2
	firstHalf := strings.Join(candidateMsgs[:mid], " ")
	secondHalf := strings.Join(candidateMsgs[mid:], " ")

	firstAvg := avgWordLength(firstHalf)
	secondAvg := avgWordLength(secondHalf)

	// значительный сдвиг средней длины = смена стиля
	diff := firstAvg - secondAvg
	if diff < 0 {
		diff = -diff
	}

	return diff > 2.0
}

// средняя длина слова в тексте
func avgWordLength(text string) float64 {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}
	totalLen := 0
	for _, w := range words {
		totalLen += len(w)
	}
	return float64(totalLen) / float64(len(words))
}

// секция античит-сигналов для промпта оценки
func buildAntiCheatSection(report *AntiCheatReport) string {
	var sb strings.Builder
	sb.WriteString("=== BEHAVIORAL SIGNALS ===\n")
	sb.WriteString(fmt.Sprintf("Total candidate messages: %d\n", report.TotalMessages))
	sb.WriteString(fmt.Sprintf("Voice messages: %d, Text messages: %d (voice ratio: %.1f%%)\n",
		report.VoiceMessageCount, report.TextMessageCount, report.VoiceToTextRatio*100))
	sb.WriteString(fmt.Sprintf("Average response time: %.1f sec\n", report.AvgResponseTimeSec))
	sb.WriteString(fmt.Sprintf("Fastest response: %d sec, Slowest: %d sec\n",
		report.MinResponseTimeSec, report.MaxResponseTimeSec))

	if report.FastResponseCount > 0 {
		sb.WriteString(fmt.Sprintf("WARNING: %d suspiciously fast responses (< 5 sec with substantial content)\n",
			report.FastResponseCount))
	}

	if len(report.Flags) > 0 {
		sb.WriteString(fmt.Sprintf("Flags: %s\n", strings.Join(report.Flags, ", ")))
	}

	return sb.String()
}
