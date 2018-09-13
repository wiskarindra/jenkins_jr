package jenkins_jr

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindOrCreateInfluencer(t *testing.T) {
	env := newEnvTest()
	defer clearDBTestData(env.DB)

	inputName := "  random  nAme. "
	expectedName := "Random Name"

	ctx := context.Background()
	tx := env.DB.MustBegin()
	influencer, err := env.findOrCreateInfluencer(ctx, tx, inputName)
	tx.Commit()

	assert.Equal(t, nil, err)
	assert.Equal(t, expectedName, influencer.Name)

	influencerID := influencer.ID
	inputName = "  Random  Name. "

	tx = env.DB.MustBegin()
	influencer, err = env.findOrCreateInfluencer(ctx, tx, inputName)
	tx.Commit()

	assert.Equal(t, nil, err)
	assert.Equal(t, influencerID, influencer.ID, "findOrCreateInfluencer should not create new influencer when it already exists")
}

func TestNormalizeInfluencerName(t *testing.T) {
	testcase := [][2]string{
		[2]string{"  random  nAme. ", "Random Name"},
		[2]string{"random name", "Random Name"},
		[2]string{"randomname", "Randomname"},
		[2]string{"Random Name", "Random Name"},
		[2]string{"Random Al-Name", "Random Alname"},
	}

	for _, ts := range testcase {
		assert.Equal(t, ts[1], normalizeInfluencerName(ts[0]))
	}
}

func TestGetInfluencerNameFromUsername(t *testing.T) {
	testcase := [][2]string{
		[2]string{"random-name", "Random Name"},
		[2]string{"random-name-", "Random Name"},
		[2]string{"random-al-name-", "Random Al Name"},
	}

	for _, ts := range testcase {
		assert.Equal(t, ts[1], getInfluencerNameFromUsername(ts[0]))
	}
}

func TestGetInfluencerUsernameFromName(t *testing.T) {
	testcase := [][2]string{
		[2]string{"Random Name", "random-name"},
		[2]string{"Random Al Name", "random-al-name"},
	}

	for _, ts := range testcase {
		assert.Equal(t, ts[1], getInfluencerUsernameFromName(ts[0]))
	}
}
