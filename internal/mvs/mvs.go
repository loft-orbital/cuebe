/*
Copyright © 2021 Loft Orbital

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package mvs

import (
	"context"

	"golang.org/x/mod/module"
	"golang.org/x/sync/errgroup"
)

type Reqs interface {
	// Required returns the module versions explicitly required by m itself.
	// The caller must not modify the returned list.
	Required(m module.Version) ([]module.Version, error)

	// Compare returns an integer comparing two module versions.
	// The result will be 0 if v == w, -1 if v < w, or +1 if v > w.
	Compare(v, w string) int
}

// BuildList executes the mvs algorithm to find the build list for root.
// smever.Compare is used to compare versions
func BuildList(root module.Version, reqs Reqs) ([]module.Version, error) {
	visited := map[module.Version]bool{}
	selected := map[string]string{}
	next := []module.Version{root}
	wc := 10                           // number of parallel worker
	rqc := make(chan []module.Version) // module requirements channel workers use to communicate

	it := 0 // current index in the 'next' slice
	for it < len(next) {
		eg, _ := errgroup.WithContext(context.Background())
		si := min(len(next), it+wc) // number of worker that will be spawned
		done := make(chan error)

		// fire workers
		for i := it; i < si; i++ {
			mod := next[i]
			eg.Go(buildWorker(mod, reqs, rqc))
			visited[mod] = true
			it++
		}
		go func() {
			done <- eg.Wait()
			close(done)
		}()

	Res:
		for {
			select {
			case r := <-rqc:
				// process new requirements received
				for _, m := range r {
					if v, ok := visited[m]; ok && v {
						continue // skip already visited node
					}
					if v, ok := selected[m.Path]; !ok || reqs.Compare(m.Version, v) > 0 {
						selected[m.Path] = m.Version
					}
					next = append(next, m)
				}
			case err := <-done:
				// jobs done, go to next iteration
				if err != nil {
					return nil, err
				}
				break Res
			}
		}
	}

	// convert the selected map to a module.Version slice
	res := make([]module.Version, 0, len(selected))
	for p, v := range selected {
		res = append(res, module.Version{Path: p, Version: v})
	}
	return res, nil
}

// buildWorker returns a worker for BuildList.
func buildWorker(mod module.Version, rg Reqs, reqs chan<- []module.Version) func() error {
	return func() error {
		r, err := rg.Required(mod)
		if err != nil {
			return err
		}

		reqs <- r
		return nil
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
