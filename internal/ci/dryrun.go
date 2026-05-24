package ci

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

type DryRunExecutor struct {
	Delay       time.Duration
	SuccessRate float64
	rng         *rand.Rand
}

func NewDryRunExecutor() *DryRunExecutor {
	return &DryRunExecutor{
		Delay:       500 * time.Millisecond,
		SuccessRate: 1.0,
		rng:         rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (e *DryRunExecutor) Execute(ctx context.Context, step StepConfig) (*StepResult, error) {
	if len(step.Command) == 0 && step.Image == "" {
		return &StepResult{
			Name:   step.Name,
			Status: StatusFailed,
			Err:    fmt.Errorf("%w: commande ou image requise", ErrInvalidStep),
		}, nil
	}

	select {
	case <-ctx.Done():
		return &StepResult{
			Name:   step.Name,
			Status: StatusCancelled,
		}, nil
	case <-time.After(e.Delay):
	}

	select {
	case <-ctx.Done():
		return &StepResult{
			Name:   step.Name,
			Status: StatusCancelled,
		}, nil
	default:
	}

	output := SimulatedOutput(step)
	duration := e.Delay

	if e.rng.Float64() >= e.SuccessRate {
		return &StepResult{
			Name:     step.Name,
			Status:   StatusFailed,
			Output:   output,
			Duration: duration,
			Err:      fmt.Errorf("dry-run: étape simulée en échec"),
		}, nil
	}

	return &StepResult{
		Name:     step.Name,
		Status:   StatusSucceeded,
		Output:   output,
		Duration: duration,
	}, nil
}

func SimulatedOutput(step StepConfig) string {
	if step.Image != "" {
		return containerRunLog(step.Image, step.Command)
	}

	cmd := strings.Join(step.Command, " ")
	cmdLower := strings.ToLower(cmd)

	switch {
	case strings.Contains(cmdLower, "go build"):
		return buildLog()
	case strings.Contains(cmdLower, "go test"):
		return testLog(15, 0, 2)
	case strings.Contains(cmdLower, "mvn clean compile"):
		return mavenBuildLog()
	case strings.Contains(cmdLower, "mvn test") || strings.Contains(cmdLower, "mvn verify"):
		return mavenTestLog()
	case strings.Contains(cmdLower, "gradle build"):
		return gradleBuildLog()
	case strings.Contains(cmdLower, "gradle test"):
		return gradleTestLog()
	case strings.Contains(cmdLower, "compose") || strings.Contains(cmdLower, "up -d"):
		return deployLog("ade-config")
	case strings.Contains(cmdLower, "npm run build") || strings.Contains(cmdLower, "npm"):
		return npmBuildLog()
	default:
		return genericCommandLog(cmd)
	}
}

func buildLog() string {
	return "> go build ./...\n" +
		"go: downloading github.com/foo/bar v1.0.0\n" +
		"go: downloading github.com/baz/qux v2.0.0\n" +
		"go: downloading golang.org/x/text v0.14.0\n" +
		"✓ build succeeded (3.24s)"
}

func testLog(passed, failed, skipped int) string {
	var b strings.Builder
	b.WriteString("> go test ./...\n")
	b.WriteString(fmt.Sprintf("ok  \tmyproject/pkg/foo\t0.342s\n"))
	b.WriteString(fmt.Sprintf("ok  \tmyproject/pkg/bar\t0.567s\n"))
	b.WriteString(fmt.Sprintf("ok  \tmyproject/pkg/baz\t0.891s\n"))
	b.WriteString(fmt.Sprintf("ok  \tmyproject/pkg/qux\t0.123s\n"))
	b.WriteString(fmt.Sprintf("✓ %d passed, %d failed, %d skipped (3.45s total)", passed, failed, skipped))
	return b.String()
}

func mavenBuildLog() string {
	return "> mvn clean compile\n" +
		"[INFO] Scanning for projects...\n" +
		"[INFO] \n" +
		"[INFO] ------------------< com.example:my-app >------------------\n" +
		"[INFO] Building my-app 1.0.0\n" +
		"[INFO]   from pom.xml\n" +
		"[INFO] \n" +
		"[INFO] --- maven-clean-plugin:3.2.0:clean (default-clean) @ my-app ---\n" +
		"[INFO] \n" +
		"[INFO] --- maven-compiler-plugin:3.11.0:compile (default-compile) @ my-app ---\n" +
		"[INFO] Compiling 42 source files to target/classes\n" +
		"[INFO] \n" +
		"[INFO] BUILD SUCCESS (5.67s)"
}

func mavenTestLog() string {
	return "> mvn test\n" +
		"[INFO] Scanning for projects...\n" +
		"[INFO] \n" +
		"[INFO] --- maven-surefire-plugin:3.1.2:test (default-test) @ my-app ---\n" +
		"[INFO] \n" +
		"[INFO] Tests run: 128, Failures: 0, Errors: 0, Skipped: 3\n" +
		"[INFO] \n" +
		"[INFO] BUILD SUCCESS (12.34s)"
}

func gradleBuildLog() string {
	return "> gradle build -x test\n" +
		"Configuration on demand is an incubating feature.\n" +
		"Downloading https://services.gradle.org/distributions/gradle-8.5-bin.zip\n" +
		"................................................................\n" +
		"Unzipping...\n" +
		"\n" +
		"> Task :compileJava UP-TO-DATE\n" +
		"> Task :processResources UP-TO-DATE\n" +
		"> Task :classes UP-TO-DATE\n" +
		"> Task :jar UP-TO-DATE\n" +
		"> Task :assemble UP-TO-DATE\n" +
		"> Task :check SKIPPED\n" +
		"> Task :build SKIPPED\n" +
		"\n" +
		"BUILD SUCCESSFUL in 8.42s"
}

func gradleTestLog() string {
	return "> gradle test\n" +
		"> Task :compileJava UP-TO-DATE\n" +
		"> Task :processResources UP-TO-DATE\n" +
		"> Task :classes UP-TO-DATE\n" +
		"> Task :compileTestJava UP-TO-DATE\n" +
		"> Task :processTestResources UP-TO-DATE\n" +
		"> Task :testClasses UP-TO-DATE\n" +
		"> Task :test\n" +
		"\n" +
		"com.example.MyServiceSpec > should create entity PASSED\n" +
		"com.example.MyServiceSpec > should update entity PASSED\n" +
		"com.example.MyServiceSpec > should delete entity PASSED\n" +
		"\n" +
		"BUILD SUCCESSFUL in 15.67s"
}

func npmBuildLog() string {
	return "> npm run build\n" +
		"\n" +
		"> my-app@1.0.0 build\n" +
		"> webpack --mode production\n" +
		"\n" +
		"assets by status 1.24 MiB [emitted]\n" +
		"  asset main.js 1.24 MiB [emitted] (name: main)\n" +
		"  asset index.html 1.02 KiB [emitted]\n" +
		"  asset styles.css 245 KiB [emitted] (name: styles)\n" +
		"orphan modules 174 KiB [orphan] 86 modules\n" +
		"runtime modules 1.49 KiB 7 modules\n" +
		"cacheable modules 1.02 MiB\n" +
		"  modules by path ./src/ 1.01 MiB 47 modules\n" +
		"  modules by path ./node_modules/ 8.62 KiB\n" +
		"webpack 5.89.0 compiled successfully in 3.45s"
}

func deployLog(serviceName string) string {
	return fmt.Sprintf("> docker compose up -d\n"+
		"[+] Running 2/2\n"+
		" ✔ Container %s  Started\n"+
		" ✔ Network ade-network   Created\n"+
		"✓ Deployment completed (5.12s)", serviceName)
}

func containerRunLog(image string, command []string) string {
	cmdStr := strings.Join(command, " ")
	shortImage := image
	if idx := strings.Index(image, ":"); idx > 0 {
		shortImage = image[:idx]
	}

	var b strings.Builder
	if len(command) > 0 {
		b.WriteString(fmt.Sprintf("> docker run --rm %s %s\n", image, cmdStr))
	} else {
		b.WriteString(fmt.Sprintf("> docker run --rm %s\n", image))
	}
	b.WriteString(fmt.Sprintf("Unable to find image '%s' locally\n", image))
	b.WriteString(fmt.Sprintf("%s: Pulling from library/%s\n", shortImage, shortImage))
	b.WriteString("Digest: sha256:abc123def456abc123def456abc123def456abc123def456abc123def456\n")
	b.WriteString("Status: Downloaded newer image for " + image + "\n")
	b.WriteString(fmt.Sprintf("✓ Container execution completed (%.2fs)", 3.0+float64(len(command))*1.5))
	return b.String()
}

func genericCommandLog(cmd string) string {
	return fmt.Sprintf("> %s\n✓ Command completed successfully (1.23s)", cmd)
}
