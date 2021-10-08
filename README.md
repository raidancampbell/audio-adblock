# audio-adblock
least common subsequence matching in audio fingerprints to remove ads and intros from your podcasts

### Current Status
 Vaguely working!  Matching isn't very good, but two passes at 30% similarity and minimum 10 second spans has yielded good results for two inputs.

## Usage
1. open `audio-adblock.go` and add the filenames of the two inputs
2. `go run audio-adblock.go` will read them in and start processing
3. data will be fingerprinted, and matching sections will be removed
4. outputs will be written to `A.mp3` and `B.mp3`, as hardcoded in `audio-adblock.go`

## to do
 - [ ] retain matched fingerprints in memory
 - [ ] serialize fingerprints for storage
 - [ ] serialize associated audio snippet for storage
 - [ ] allow more than two inputs for fingerprinting, limit to one output
 - [ ] add a halfhearted CLI
 - [ ] switch to a multithreaded LAME encoder, or ditch LAME entirely