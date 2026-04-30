// Fix: lodash getRawTag tries to temporarily override Symbol.toStringTag on DataView,
// but DataView.prototype[Symbol.toStringTag] is read-only in ES module strict mode.
// Make it writable before any lodash code runs.
const desc = Object.getOwnPropertyDescriptor(DataView.prototype, Symbol.toStringTag)
if (desc && !desc.writable) {
  Object.defineProperty(DataView.prototype, Symbol.toStringTag, {
    value: 'DataView',
    writable: true,
    enumerable: false,
    configurable: true,
  })
}
