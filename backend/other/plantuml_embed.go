package other

import (
	_ "embed"
	"os"
	"path/filepath"
	"sync"
)

//go:embed assets/plantuml.jar
var embeddedPlantUMLJar []byte

var (
	plantumlJarOnce sync.Once
	plantumlJarPath string
	plantumlJarErr  error
)

func ensureEmbeddedPlantUMLJar() (string, error) {
	plantumlJarOnce.Do(func() {
		dir := filepath.Join(os.TempDir(), "ctools")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			plantumlJarErr = err
			return
		}
		target := filepath.Join(dir, "plantuml.jar")
		if err := os.WriteFile(target, embeddedPlantUMLJar, 0o755); err != nil {
			plantumlJarErr = err
			return
		}
		plantumlJarPath = target
	})
	return plantumlJarPath, plantumlJarErr
}
