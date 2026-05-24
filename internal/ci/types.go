package ci

type StageType string

const (
	StageBuild           StageType = "build"
	StageUnitTest        StageType = "unit-test"
	StageIntegrationTest StageType = "integration-test"
	StageTestDeploy      StageType = "test-deploy"
	StageE2E             StageType = "e2e"
	StagePreprod         StageType = "preprod"
)

func AllStages() []StageType {
	return []StageType{
		StageBuild,
		StageUnitTest,
		StageIntegrationTest,
		StageTestDeploy,
		StageE2E,
		StagePreprod,
	}
}

var stageDescriptions = map[StageType]string{
	StageBuild:           "Construction du projet",
	StageUnitTest:        "Tests unitaires",
	StageIntegrationTest: "Tests d'intégration",
	StageTestDeploy:      "Déploiement en environnement de test",
	StageE2E:             "Tests end-to-end",
	StagePreprod:         "Déploiement en préproduction",
}

func StageDescription(t StageType) string {
	return stageDescriptions[t]
}

type StageStatus string

const (
	StatusPending   StageStatus = "pending"
	StatusRunning   StageStatus = "running"
	StatusSucceeded StageStatus = "succeeded"
	StatusFailed    StageStatus = "failed"
	StatusSkipped   StageStatus = "skipped"
	StatusCancelled StageStatus = "cancelled"
)

type StepConfig struct {
	Name    string            `yaml:"name" json:"name"`
	Image   string            `yaml:"image,omitempty" json:"image,omitempty"`
	Command []string          `yaml:"command,omitempty" json:"command,omitempty"`
	Env     map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
	WorkDir string            `yaml:"workdir,omitempty" json:"workdir,omitempty"`
}

type StageConfig struct {
	Type    StageType    `yaml:"type" json:"type"`
	Name    string       `yaml:"name,omitempty" json:"name,omitempty"`
	Steps   []StepConfig `yaml:"steps" json:"steps"`
	Enabled bool         `yaml:"enabled" json:"enabled"`
}

type PipelineConfig struct {
	Stages []StageConfig `yaml:"stages" json:"stages"`
}
