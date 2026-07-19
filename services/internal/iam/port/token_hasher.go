package port

// TokenHasher hashes tokens for secure storage.
// Used by the use-case layer to hash refresh tokens before persisting them.
type TokenHasher interface {
	HashToken(token string) string
}
