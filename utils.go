package planet

func validSequence(old uint16, new uint16) bool {
	return ((new > old) && (new-old <= 32768)) || ((new < old) && (old-new > 32768))

}
