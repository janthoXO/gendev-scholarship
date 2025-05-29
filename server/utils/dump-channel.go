package utils

// takes a channel and forwards all messages into the void
func DumpChannel[T any](ch <-chan T) {
	go func() {
		for range ch {
			// This function is intentionally left empty.
			// It serves to keep the channel open and prevent blocking.
			// You can add logging or other operations here if needed.
		}
	}()
}
