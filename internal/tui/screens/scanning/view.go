package scanning

func (s ScanningScreen) View() string {
	return s.Spinner.View() + " Escaneando el repositorio remoto..."
}
