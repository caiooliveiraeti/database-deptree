package java

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestAnalyzeJavaFiles(t *testing.T) {
	// Setup temporary directory with Java files for testing
	tmpDir, err := ioutil.TempDir("", "java_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

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
	ioutil.WriteFile(tmpDir+"/TestEntity.java", []byte(javaFileContent), 0644)

	javaAnalyzer := JavaAnalyzer{RootDir: tmpDir}
	deps, err := javaAnalyzer.Analyze()
	if err != nil {
		t.Fatalf("Failed to analyze Java files: %v", err)
	}

	if len(deps) == 0 {
		t.Fatalf("Expected dependencies to be found, got %d", len(deps))
	}
}
