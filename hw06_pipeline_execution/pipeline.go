package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	for _, stage := range stages {
		ch := make(Bi)
		go func(in In) {
			defer close(ch)
			for v := range in {
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
		}(in)
		in = stage(ch)
	}

	return in
}
