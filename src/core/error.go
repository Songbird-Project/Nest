package core

import "github.com/charmbracelet/lipgloss/v2"

func Info(text string) string {
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Blue)
	return infoStyle.Render(text)
}

func Warn(text string) string {
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Yellow)
	return warnStyle.Render(text)
}

func Error(text string) string {
	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Red)
	return errorStyle.Render(text)
}
