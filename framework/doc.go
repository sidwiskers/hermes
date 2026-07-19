// Package framework implements Hermes routing, filters, middleware, and
// handler contexts without owning transport or runtime lifecycle.
//
// Most applications use these declarations through the root hermes facade.
// Import framework directly only when building an independent router or
// framework integration around the standalone api client.
package framework
