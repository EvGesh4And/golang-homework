package hw06pipelineexecution

type (
	In  = <-chan interface{}
	Out = In
	Bi  = chan interface{}
)

type Stage func(in In) (out Out)

func ExecutePipeline(in In, done In, stages ...Stage) Out {
	checker := func(in In) Out {
		out := make(Bi)

		go func() {
			defer func() {
				close(out)
				<-in // Считываем значение из канала, чтобы stage мог завершиться
				// Считываем либо результат работы stage, либо из закрытого канала
			}()

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
					case out <- v:
					}
				}
			}
		}()

		return out
	}

	for _, stage := range stages {
		in = stage(checker(in))
	}

	out := checker(in)

	return out
}
