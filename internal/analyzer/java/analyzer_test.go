package java

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/caiooliveiraeti/database-deptree/internal/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const ownerFixture = `
@Entity
@Table(name = "owners")
public class Owner {
    @Id
    private Long id;

    @OneToMany(cascade = CascadeType.ALL, mappedBy = "owner")
    private Set<Pet> pets;
}
`

const petFixture = `
@Entity
@Table(name = "pets")
public class Pet {
    @Id
    private Long id;

    @ManyToOne
    @JoinColumn(name = "owner_id")
    private Owner owner;

    @ManyToOne
    @JoinColumn(name = "type_id")
    private PetType type;
}
`

const repoFixture = `
public interface OwnerRepository extends JpaRepository<Owner, Long> {
    @Query("SELECT o FROM Owner o WHERE o.id = ?1")
    List<Owner> findById(Long id);

    @Procedure("calculate_tax")
    BigDecimal computeTax(Long id);
}
`

const nativeRepoFixture = `
public interface OwnerRepository extends JpaRepository<Owner, Long> {
    @Query(value = "SELECT * FROM owners WHERE id = ?1", nativeQuery = true)
    List<Owner> findByIdNative(Long id);
}
`

func TestAnalyze_StoredIn(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Owner.java"), []byte(ownerFixture), 0644))

	edges, err := New(tmpDir, "").Analyze(context.Background())
	require.NoError(t, err)

	rels := relSet(edges)
	assert.True(t, rels["STORED_IN"], "expected STORED_IN (ENTITY -> TABLE)")
}

func TestAnalyze_JoinEdges(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Owner.java"), []byte(ownerFixture), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Pet.java"), []byte(petFixture), 0644))

	edges, err := New(tmpDir, "").Analyze(context.Background())
	require.NoError(t, err)

	rels := relSet(edges)
	assert.True(t, rels["MANY_TO_ONE"], "expected MANY_TO_ONE (pets -> owners, pets -> types)")
	assert.True(t, rels["ONE_TO_MANY"], "expected ONE_TO_MANY (owners -> pets)")

	// Verify that Pet's ManyToOne to Owner resolves to table "owners" (not class name)
	for _, e := range edges {
		if e.Relationship == "MANY_TO_ONE" && e.Source.ID == "TABLE:pets" {
			if e.Properties["toClass"] == "Owner" {
				assert.Equal(t, "TABLE:owners", e.Target.ID,
					"Pet @ManyToOne Owner should resolve to TABLE:owners")
			}
		}
	}
}

func TestAnalyze_Repository(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Owner.java"), []byte(ownerFixture), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "OwnerRepository.java"), []byte(repoFixture), 0644))

	edges, err := New(tmpDir, "").Analyze(context.Background())
	require.NoError(t, err)

	rels := relSet(edges)
	assert.True(t, rels["MANAGES"])
	assert.True(t, rels["QUERIES"])
	assert.True(t, rels["CALLS"])

	// JPQL FROM Owner must resolve to TABLE:owners (not a TABLE node named "Owner")
	assert.True(t, rels["USES_TABLE"], "JPQL query should resolve entity to USES_TABLE")
	assert.False(t, rels["USES_ENTITY"], "JPQL query must not produce USES_ENTITY")
	for _, e := range edges {
		if e.Relationship == "USES_TABLE" {
			assert.Equal(t, "TABLE:owners", e.Target.ID, "JPQL FROM Owner must resolve to TABLE:owners")
		}
	}
}

func TestAnalyze_NativeQuery(t *testing.T) {
	tmpDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "Owner.java"), []byte(ownerFixture), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "OwnerRepository.java"), []byte(nativeRepoFixture), 0644))

	edges, err := New(tmpDir, "").Analyze(context.Background())
	require.NoError(t, err)

	rels := relSet(edges)
	assert.True(t, rels["USES_TABLE"], "native query should produce USES_TABLE")
	for _, e := range edges {
		if e.Relationship == "USES_TABLE" {
			assert.Equal(t, "TABLE:owners", e.Target.ID, "native FROM owners must point to TABLE:owners")
		}
	}
}

func TestAnalyze_EmptyDir(t *testing.T) {
	edges, err := New(t.TempDir(), "").Analyze(context.Background())
	require.NoError(t, err)
	assert.Empty(t, edges)
}

func TestNodeID_ProcedureCrossAnalyzer(t *testing.T) {
	tmpDir := t.TempDir()
	content := `
@Entity
public class MyEntity {}
public interface MyRepo extends JpaRepository<MyEntity, Long> {
    @Procedure("calculate_tax")
    void run();
}
`
	require.NoError(t, os.WriteFile(filepath.Join(tmpDir, "MyEntity.java"), []byte(content), 0644))

	edges, err := New(tmpDir, "").Analyze(context.Background())
	require.NoError(t, err)

	for _, e := range edges {
		if e.Relationship == "CALLS" {
			assert.Equal(t, "PROCEDURE:calculate_tax", e.Target.ID,
				"must match Oracle's NodeID for the same procedure")
		}
	}
}

func relSet(edges []graph.Edge) map[string]bool {
	m := make(map[string]bool)
	for _, e := range edges {
		m[e.Relationship] = true
	}
	return m
}
