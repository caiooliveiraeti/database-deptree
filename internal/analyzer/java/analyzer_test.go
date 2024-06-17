package java

import (
	"os"
	"testing"
)

func TestAnalyzeJavaFiles(t *testing.T) {
	tmpDir := t.TempDir()

	javaFileContent := `
        @Entity
        @Table(name = "TestTable")
        public class TestEntity {
            @Id
            private Long id;
        }

        public interface TestRepository extends JpaRepository<TestEntity, Long> {
            @Query("SELECT t FROM TestEntity t WHERE t.id = ?1")
            List<TestEntity> findById(Long id);
        }
    `

	err := writeFile(tmpDir+"/TestEntity.java", javaFileContent)
	if err != nil {
		t.Fatalf("Failed to analyze Java files: %v", err)
	}

	javaAnalyzer := JavaAnalyzer{RootDir: tmpDir}
	deps, err := javaAnalyzer.Analyze()
	if err != nil {
		t.Fatalf("Failed to analyze Java files: %v", err)
	}

	if len(deps) == 0 {
		t.Fatalf("Expected dependencies to be found, got %d", len(deps))
	}
}

func writeFile(path string, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}
