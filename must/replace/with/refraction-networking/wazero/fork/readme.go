package fork

/************************************************************
 *  This file is a dummy file to keep this path importable.
 *         No functional code should be added here.
 ************************************************************/

// In this forked version of tetratelabs/wazero, we have
// updated some of the existing implementations and added some
// new features to better support the purpose of using
// WebAssembly in the context of networking, especially for
// application-layer transport protocols with tunneling
// capabilities.
//
// To maintain the maximum compatibility with the upstream tetratelabs/wazero,
// we have kept the original package import path as "github.com/tetratelabs/wazero"
// to prevent future merge conflicts when pulling the latest changes from the upstream.
// However, this does prevent any user from directly importing the forked version
// at "github.com/refraction-networking/wazero" without using the replace directive in go.mod.
//
// Before we can figure out a better way to handle this issue, we added a dummy package here
// with no functional code within, in order to detect if the user is importing our fork
// (with this dummy package) correctly with a replace directive or the upstream version
// without using a replace directive.
//
// For each upstream tag vX.Y.Z of wazero, we release a tag vX.Y.Z-w to indicate the patched/modified
// version of the upstream tag vX.Y.Z. This is to help users understand the compatibility of our fork
// with the upstream version.
//
// We try our best to keep the maximum backward compatibility in the future versions of our fork,
// but sometimes it might not be possible if our design drastically changes to better support
// our use case. Therefore, we recommend indirect dependents (e.g., your project depends on a project
// that depends on this fork) to use the replace directive with the exact version of our fork used by
// your dependency.
