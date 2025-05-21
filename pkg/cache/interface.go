package cache

type Cache interface {
	Update(vmiName string, status VmiStatus) bool
	Delete(vmiName string)
}
