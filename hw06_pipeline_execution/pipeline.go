package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {

	checker := func(in In, ch Bi) {
		defer close(ch)
		for {
			select {
			case <-done:
				return
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case <-done:
					return
				default:
					select {
					case <-done:
						return
					case ch <- v:
					}
				}
			}
		}
	}

	for _, stage := range stages {
		ch := make(Bi)
		go checker(in, ch)
		in = stage(ch)
	}

	ch := make(Bi)

	go checker(in, ch)

	return ch
}
