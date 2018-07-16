package domain_test

import (
	"testing"

	"github.com/stretchr/testify/assert" // More readable test assertions.

	"github.com/marccarre/kubernetes-deployment-strategies-workload/pkg/domain"
)

var user = domain.User{
	FirstName:  "Luke",
	FamilyName: "Skywalker",
	Age:        20,
}

func TestFullNameShouldConcatenateFirstAndFamilyName(t *testing.T) {
	assert.Equal(t, "Luke Skywalker", user.FullName())
}

func TestMarshalShouldReturnUserAsJSON(t *testing.T) {
	bytes, err := user.Marshal()
	assert.NoError(t, err)
	assert.Equal(t, "{\"firstName\":\"Luke\",\"familyName\":\"Skywalker\",\"age\":20}", string(bytes))
}

func TestUnmarshalShouldReturnJSONAsUser(t *testing.T) {
	user, err := domain.UnmarshalUser([]byte("{\"id\":1337,\"firstName\":\"Foo\",\"familyName\":\"Bar\"}"))
	assert.NoError(t, err)
	assert.Equal(t, domain.User{
		ID:         1337,
		FirstName:  "Foo",
		FamilyName: "Bar",
		Age:        0,
	}, *user)
}

func TestUnmarshalInvalidJSONShouldReturnError(t *testing.T) {
	user, err := domain.UnmarshalUser([]byte("not-valid-json"))
	assert.EqualError(t, err, "invalid character 'o' in literal null (expecting 'u')")
	assert.Nil(t, user)
}

func TestUnmarshalInvalidJSONUserShouldReturnError(t *testing.T) {
	user, err := domain.UnmarshalUser([]byte("{\"foo\":\"bar\"}"))
	assert.EqualError(t, err, "invalid JSON: doesn't yield a valid user: {\"foo\":\"bar\"}")
	assert.Nil(t, user)
}
