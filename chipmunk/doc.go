// Chipmunk-GO binding for chipmunk physics engine
//
// See chipmunk-physics.net for chipmunk source-code
//
// Just like the original code, this binding is released
// under the MIT license. See LICENSE for more details.
//
// IMPORTANT!
// The current code might leak memory since chipmunk allocated
// objects need to be manually free'd. You should use defer obj.Free()
// but the order is important!
//
// * free all shapes
// * free all bodies
// * free the space object
//
// This is the opposite of the creation order and therefor you can't use
// defer the usual way (at least not now).
package chipmunk
