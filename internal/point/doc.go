/*
Package point provides convenience logic for manipulating points and regions in
2-space.

NOTE: this package mostly exists because I (Josh) had never used the core
"image" package much, and didn't know about its Point/Rectangle structs; one
point of view would say that many of the utililyt methods should just be broken
out (maybe into "bork/internal/moremath") as function-taking-data (rather than
receiving). The Point structure is (coincidentally) cast-compatible with the
standard one in the "image" package; the Box isn't becasue of Go's conservatism
composing poorly with my short sightedness.

TODO: de-conflict and align better with core "image".Point et al along
spiritually identical or similar methods.

*/
package point
