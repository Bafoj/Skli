package scanning

func (s ScanningScreen) View() string {
	return s.Spinner.View() + " Scanning remote repository..."
}
