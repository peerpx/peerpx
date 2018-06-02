package config

const (
	confBadSyntax = "foobartoto"
	confGood      = `string:bar
# comment
int:7
float:7.12
bool: true
`
)

// TODO faire les tests de config.basic

/*
func TestBasicConfig(t *testing.T) {
	// bad syntax
	assert.Error(t, InitBasicConfig(strings.NewReader(confBadSyntax)))

	// good syntax
	if assert.NoError(t, InitBasicConfig(strings.NewReader(confGood))) {
		i := Get("string")
		assert.Equal(t, interface{}("bar"), i)

		// not found -> panic
		assert.Panics(t, func() { GetOrPanic("unicorn") })

		// GetString
		s := GetString("string")
		assert.IsType(t, "foo", s)
		assert.Equal(t, "bar", s)

		s = GetString("int")
		assert.IsType(t, "foo", s)
		assert.Equal(t, "7", s)

		s = GetString("float")
		assert.IsType(t, "foo", s)
		assert.Equal(t, "7.12", s)

		// GetFloat64
		assert.Equal(t, 0.00, GetFloat64("string"))
		assert.Panics(t, func() { GetFloat64OrPanic("string") })
		assert.Equal(t, 7.00, GetFloat64("int"))
		assert.Equal(t, 7.12, GetFloat64("float"))
		assert.Equal(t, 7.12, GetFloat64OrPanic("float"))
	}
}
*/
