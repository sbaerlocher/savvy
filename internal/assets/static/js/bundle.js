// packages/alpinejs/src/scheduler.js
var flushPending = false;
var flushing = false;
var queue = [];
var lastFlushedIndex = -1;
function scheduler(callback) {
  queueJob(callback);
}
function queueJob(job) {
  if (!queue.includes(job))
    queue.push(job);
  queueFlush();
}
function dequeueJob(job) {
  let index = queue.indexOf(job);
  if (index !== -1 && index > lastFlushedIndex)
    queue.splice(index, 1);
}
function queueFlush() {
  if (!flushing && !flushPending) {
    flushPending = true;
    queueMicrotask(flushJobs);
  }
}
function flushJobs() {
  flushPending = false;
  flushing = true;
  for (let i = 0; i < queue.length; i++) {
    queue[i]();
    lastFlushedIndex = i;
  }
  queue.length = 0;
  lastFlushedIndex = -1;
  flushing = false;
}

// packages/alpinejs/src/reactivity.js
var reactive;
var effect;
var release;
var raw;
var shouldSchedule = true;
function disableEffectScheduling(callback) {
  shouldSchedule = false;
  callback();
  shouldSchedule = true;
}
function setReactivityEngine(engine) {
  reactive = engine.reactive;
  release = engine.release;
  effect = (callback) => engine.effect(callback, { scheduler: (task) => {
    if (shouldSchedule) {
      scheduler(task);
    } else {
      task();
    }
  } });
  raw = engine.raw;
}
function overrideEffect(override) {
  effect = override;
}
function elementBoundEffect(el) {
  let cleanup2 = () => {
  };
  let wrappedEffect = (callback) => {
    let effectReference = effect(callback);
    if (!el._x_effects) {
      el._x_effects = /* @__PURE__ */ new Set();
      el._x_runEffects = () => {
        el._x_effects.forEach((i) => i());
      };
    }
    el._x_effects.add(effectReference);
    cleanup2 = () => {
      if (effectReference === void 0)
        return;
      el._x_effects.delete(effectReference);
      release(effectReference);
    };
    return effectReference;
  };
  return [wrappedEffect, () => {
    cleanup2();
  }];
}
function watch(getter, callback) {
  let firstTime = true;
  let oldValue;
  let effectReference = effect(() => {
    let value = getter();
    JSON.stringify(value);
    if (!firstTime) {
      queueMicrotask(() => {
        callback(value, oldValue);
        oldValue = value;
      });
    } else {
      oldValue = value;
    }
    firstTime = false;
  });
  return () => release(effectReference);
}

// packages/alpinejs/src/mutation.js
var onAttributeAddeds = [];
var onElRemoveds = [];
var onElAddeds = [];
function onElAdded(callback) {
  onElAddeds.push(callback);
}
function onElRemoved(el, callback) {
  if (typeof callback === "function") {
    if (!el._x_cleanups)
      el._x_cleanups = [];
    el._x_cleanups.push(callback);
  } else {
    callback = el;
    onElRemoveds.push(callback);
  }
}
function onAttributesAdded(callback) {
  onAttributeAddeds.push(callback);
}
function onAttributeRemoved(el, name, callback) {
  if (!el._x_attributeCleanups)
    el._x_attributeCleanups = {};
  if (!el._x_attributeCleanups[name])
    el._x_attributeCleanups[name] = [];
  el._x_attributeCleanups[name].push(callback);
}
function cleanupAttributes(el, names) {
  if (!el._x_attributeCleanups)
    return;
  Object.entries(el._x_attributeCleanups).forEach(([name, value]) => {
    if (names === void 0 || names.includes(name)) {
      value.forEach((i) => i());
      delete el._x_attributeCleanups[name];
    }
  });
}
function cleanupElement(el) {
  el._x_effects?.forEach(dequeueJob);
  while (el._x_cleanups?.length)
    el._x_cleanups.pop()();
}
var observer = new MutationObserver(onMutate);
var currentlyObserving = false;
function startObservingMutations() {
  observer.observe(document, { subtree: true, childList: true, attributes: true, attributeOldValue: true });
  currentlyObserving = true;
}
function stopObservingMutations() {
  flushObserver();
  observer.disconnect();
  currentlyObserving = false;
}
var queuedMutations = [];
function flushObserver() {
  let records = observer.takeRecords();
  queuedMutations.push(() => records.length > 0 && onMutate(records));
  let queueLengthWhenTriggered = queuedMutations.length;
  queueMicrotask(() => {
    if (queuedMutations.length === queueLengthWhenTriggered) {
      while (queuedMutations.length > 0)
        queuedMutations.shift()();
    }
  });
}
function mutateDom(callback) {
  if (!currentlyObserving)
    return callback();
  stopObservingMutations();
  let result = callback();
  startObservingMutations();
  return result;
}
var isCollecting = false;
var deferredMutations = [];
function deferMutations() {
  isCollecting = true;
}
function flushAndStopDeferringMutations() {
  isCollecting = false;
  onMutate(deferredMutations);
  deferredMutations = [];
}
function onMutate(mutations) {
  if (isCollecting) {
    deferredMutations = deferredMutations.concat(mutations);
    return;
  }
  let addedNodes = [];
  let removedNodes = /* @__PURE__ */ new Set();
  let addedAttributes = /* @__PURE__ */ new Map();
  let removedAttributes = /* @__PURE__ */ new Map();
  for (let i = 0; i < mutations.length; i++) {
    if (mutations[i].target._x_ignoreMutationObserver)
      continue;
    if (mutations[i].type === "childList") {
      mutations[i].removedNodes.forEach((node) => {
        if (node.nodeType !== 1)
          return;
        if (!node._x_marker)
          return;
        removedNodes.add(node);
      });
      mutations[i].addedNodes.forEach((node) => {
        if (node.nodeType !== 1)
          return;
        if (removedNodes.has(node)) {
          removedNodes.delete(node);
          return;
        }
        if (node._x_marker)
          return;
        addedNodes.push(node);
      });
    }
    if (mutations[i].type === "attributes") {
      let el = mutations[i].target;
      let name = mutations[i].attributeName;
      let oldValue = mutations[i].oldValue;
      let add2 = () => {
        if (!addedAttributes.has(el))
          addedAttributes.set(el, []);
        addedAttributes.get(el).push({ name, value: el.getAttribute(name) });
      };
      let remove = () => {
        if (!removedAttributes.has(el))
          removedAttributes.set(el, []);
        removedAttributes.get(el).push(name);
      };
      if (el.hasAttribute(name) && oldValue === null) {
        add2();
      } else if (el.hasAttribute(name)) {
        remove();
        add2();
      } else {
        remove();
      }
    }
  }
  removedAttributes.forEach((attrs, el) => {
    cleanupAttributes(el, attrs);
  });
  addedAttributes.forEach((attrs, el) => {
    onAttributeAddeds.forEach((i) => i(el, attrs));
  });
  for (let node of removedNodes) {
    if (addedNodes.some((i) => i.contains(node)))
      continue;
    onElRemoveds.forEach((i) => i(node));
  }
  for (let node of addedNodes) {
    if (!node.isConnected)
      continue;
    onElAddeds.forEach((i) => i(node));
  }
  addedNodes = null;
  removedNodes = null;
  addedAttributes = null;
  removedAttributes = null;
}

// packages/alpinejs/src/scope.js
function scope(node) {
  return mergeProxies(closestDataStack(node));
}
function addScopeToNode(node, data2, referenceNode) {
  node._x_dataStack = [data2, ...closestDataStack(referenceNode || node)];
  return () => {
    node._x_dataStack = node._x_dataStack.filter((i) => i !== data2);
  };
}
function closestDataStack(node) {
  if (node._x_dataStack)
    return node._x_dataStack;
  if (typeof ShadowRoot === "function" && node instanceof ShadowRoot) {
    return closestDataStack(node.host);
  }
  if (!node.parentNode) {
    return [];
  }
  return closestDataStack(node.parentNode);
}
function mergeProxies(objects) {
  return new Proxy({ objects }, mergeProxyTrap);
}
var mergeProxyTrap = {
  ownKeys({ objects }) {
    return Array.from(
      new Set(objects.flatMap((i) => Object.keys(i)))
    );
  },
  has({ objects }, name) {
    if (name == Symbol.unscopables)
      return false;
    return objects.some(
      (obj) => Object.prototype.hasOwnProperty.call(obj, name) || Reflect.has(obj, name)
    );
  },
  get({ objects }, name, thisProxy) {
    if (name == "toJSON")
      return collapseProxies;
    return Reflect.get(
      objects.find(
        (obj) => Reflect.has(obj, name)
      ) || {},
      name,
      thisProxy
    );
  },
  set({ objects }, name, value, thisProxy) {
    const target = objects.find(
      (obj) => Object.prototype.hasOwnProperty.call(obj, name)
    ) || objects[objects.length - 1];
    const descriptor = Object.getOwnPropertyDescriptor(target, name);
    if (descriptor?.set && descriptor?.get)
      return descriptor.set.call(thisProxy, value) || true;
    return Reflect.set(target, name, value);
  }
};
function collapseProxies() {
  let keys = Reflect.ownKeys(this);
  return keys.reduce((acc, key) => {
    acc[key] = Reflect.get(this, key);
    return acc;
  }, {});
}

// packages/alpinejs/src/interceptor.js
function initInterceptors(data2) {
  let isObject2 = (val) => typeof val === "object" && !Array.isArray(val) && val !== null;
  let recurse = (obj, basePath = "") => {
    Object.entries(Object.getOwnPropertyDescriptors(obj)).forEach(([key, { value, enumerable }]) => {
      if (enumerable === false || value === void 0)
        return;
      if (typeof value === "object" && value !== null && value.__v_skip)
        return;
      let path = basePath === "" ? key : `${basePath}.${key}`;
      if (typeof value === "object" && value !== null && value._x_interceptor) {
        obj[key] = value.initialize(data2, path, key);
      } else {
        if (isObject2(value) && value !== obj && !(value instanceof Element)) {
          recurse(value, path);
        }
      }
    });
  };
  return recurse(data2);
}
function interceptor(callback, mutateObj = () => {
}) {
  let obj = {
    initialValue: void 0,
    _x_interceptor: true,
    initialize(data2, path, key) {
      return callback(this.initialValue, () => get(data2, path), (value) => set(data2, path, value), path, key);
    }
  };
  mutateObj(obj);
  return (initialValue) => {
    if (typeof initialValue === "object" && initialValue !== null && initialValue._x_interceptor) {
      let initialize = obj.initialize.bind(obj);
      obj.initialize = (data2, path, key) => {
        let innerValue = initialValue.initialize(data2, path, key);
        obj.initialValue = innerValue;
        return initialize(data2, path, key);
      };
    } else {
      obj.initialValue = initialValue;
    }
    return obj;
  };
}
function get(obj, path) {
  return path.split(".").reduce((carry, segment) => carry[segment], obj);
}
function set(obj, path, value) {
  if (typeof path === "string")
    path = path.split(".");
  if (path.length === 1)
    obj[path[0]] = value;
  else if (path.length === 0)
    throw error;
  else {
    if (obj[path[0]])
      return set(obj[path[0]], path.slice(1), value);
    else {
      obj[path[0]] = {};
      return set(obj[path[0]], path.slice(1), value);
    }
  }
}

// packages/alpinejs/src/magics.js
var magics = {};
function magic(name, callback) {
  magics[name] = callback;
}
function injectMagics(obj, el) {
  let memoizedUtilities = getUtilities(el);
  Object.entries(magics).forEach(([name, callback]) => {
    Object.defineProperty(obj, `$${name}`, {
      get() {
        return callback(el, memoizedUtilities);
      },
      enumerable: false
    });
  });
  return obj;
}
function getUtilities(el) {
  let [utilities, cleanup2] = getElementBoundUtilities(el);
  let utils = { interceptor, ...utilities };
  onElRemoved(el, cleanup2);
  return utils;
}

// packages/alpinejs/src/utils/error.js
function tryCatch(el, expression, callback, ...args) {
  try {
    return callback(...args);
  } catch (e) {
    handleError(e, el, expression);
  }
}
function handleError(...args) {
  return errorHandler(...args);
}
var errorHandler = normalErrorHandler;
function setErrorHandler(handler4) {
  errorHandler = handler4;
}
function normalErrorHandler(error2, el, expression = void 0) {
  error2 = Object.assign(
    error2 ?? { message: "No error message given." },
    { el, expression }
  );
  console.warn(`Alpine Expression Error: ${error2.message}

${expression ? 'Expression: "' + expression + '"\n\n' : ""}`, el);
  setTimeout(() => {
    throw error2;
  }, 0);
}

// packages/alpinejs/src/evaluator.js
var shouldAutoEvaluateFunctions = true;
function dontAutoEvaluateFunctions(callback) {
  let cache = shouldAutoEvaluateFunctions;
  shouldAutoEvaluateFunctions = false;
  let result = callback();
  shouldAutoEvaluateFunctions = cache;
  return result;
}
function evaluate(el, expression, extras = {}) {
  let result;
  evaluateLater(el, expression)((value) => result = value, extras);
  return result;
}
function evaluateLater(...args) {
  return theEvaluatorFunction(...args);
}
var theEvaluatorFunction = normalEvaluator;
function setEvaluator(newEvaluator) {
  theEvaluatorFunction = newEvaluator;
}
var theRawEvaluatorFunction;
function setRawEvaluator(newEvaluator) {
  theRawEvaluatorFunction = newEvaluator;
}
function normalEvaluator(el, expression) {
  let overriddenMagics = {};
  injectMagics(overriddenMagics, el);
  let dataStack = [overriddenMagics, ...closestDataStack(el)];
  let evaluator = typeof expression === "function" ? generateEvaluatorFromFunction(dataStack, expression) : generateEvaluatorFromString(dataStack, expression, el);
  return tryCatch.bind(null, el, expression, evaluator);
}
function generateEvaluatorFromFunction(dataStack, func) {
  return (receiver = () => {
  }, { scope: scope2 = {}, params = [], context } = {}) => {
    if (!shouldAutoEvaluateFunctions) {
      runIfTypeOfFunction(receiver, func, mergeProxies([scope2, ...dataStack]), params);
      return;
    }
    let result = func.apply(mergeProxies([scope2, ...dataStack]), params);
    runIfTypeOfFunction(receiver, result);
  };
}
var evaluatorMemo = {};
function generateFunctionFromString(expression, el) {
  if (evaluatorMemo[expression]) {
    return evaluatorMemo[expression];
  }
  let AsyncFunction = Object.getPrototypeOf(async function() {
  }).constructor;
  let rightSideSafeExpression = /^[\n\s]*if.*\(.*\)/.test(expression.trim()) || /^(let|const)\s/.test(expression.trim()) ? `(async()=>{ ${expression} })()` : expression;
  const safeAsyncFunction = () => {
    try {
      let func2 = new AsyncFunction(
        ["__self", "scope"],
        `with (scope) { __self.result = ${rightSideSafeExpression} }; __self.finished = true; return __self.result;`
      );
      Object.defineProperty(func2, "name", {
        value: `[Alpine] ${expression}`
      });
      return func2;
    } catch (error2) {
      handleError(error2, el, expression);
      return Promise.resolve();
    }
  };
  let func = safeAsyncFunction();
  evaluatorMemo[expression] = func;
  return func;
}
function generateEvaluatorFromString(dataStack, expression, el) {
  let func = generateFunctionFromString(expression, el);
  return (receiver = () => {
  }, { scope: scope2 = {}, params = [], context } = {}) => {
    func.result = void 0;
    func.finished = false;
    let completeScope = mergeProxies([scope2, ...dataStack]);
    if (typeof func === "function") {
      let promise = func.call(context, func, completeScope).catch((error2) => handleError(error2, el, expression));
      if (func.finished) {
        runIfTypeOfFunction(receiver, func.result, completeScope, params, el);
        func.result = void 0;
      } else {
        promise.then((result) => {
          runIfTypeOfFunction(receiver, result, completeScope, params, el);
        }).catch((error2) => handleError(error2, el, expression)).finally(() => func.result = void 0);
      }
    }
  };
}
function runIfTypeOfFunction(receiver, value, scope2, params, el) {
  if (shouldAutoEvaluateFunctions && typeof value === "function") {
    let result = value.apply(scope2, params);
    if (result instanceof Promise) {
      result.then((i) => runIfTypeOfFunction(receiver, i, scope2, params)).catch((error2) => handleError(error2, el, value));
    } else {
      receiver(result);
    }
  } else if (typeof value === "object" && value instanceof Promise) {
    value.then((i) => receiver(i));
  } else {
    receiver(value);
  }
}
function evaluateRaw(...args) {
  return theRawEvaluatorFunction(...args);
}
function normalRawEvaluator(el, expression, extras = {}) {
  let overriddenMagics = {};
  injectMagics(overriddenMagics, el);
  let dataStack = [overriddenMagics, ...closestDataStack(el)];
  let scope2 = mergeProxies([extras.scope ?? {}, ...dataStack]);
  let params = extras.params ?? [];
  if (expression.includes("await")) {
    let AsyncFunction = Object.getPrototypeOf(async function() {
    }).constructor;
    let rightSideSafeExpression = /^[\n\s]*if.*\(.*\)/.test(expression.trim()) || /^(let|const)\s/.test(expression.trim()) ? `(async()=>{ ${expression} })()` : expression;
    let func = new AsyncFunction(
      ["scope"],
      `with (scope) { let __result = ${rightSideSafeExpression}; return __result }`
    );
    let result = func.call(extras.context, scope2);
    return result;
  } else {
    let rightSideSafeExpression = /^[\n\s]*if.*\(.*\)/.test(expression.trim()) || /^(let|const)\s/.test(expression.trim()) ? `(()=>{ ${expression} })()` : expression;
    let func = new Function(
      ["scope"],
      `with (scope) { let __result = ${rightSideSafeExpression}; return __result }`
    );
    let result = func.call(extras.context, scope2);
    if (typeof result === "function" && shouldAutoEvaluateFunctions) {
      return result.apply(scope2, params);
    }
    return result;
  }
}

// packages/alpinejs/src/directives.js
var prefixAsString = "x-";
function prefix(subject = "") {
  return prefixAsString + subject;
}
function setPrefix(newPrefix) {
  prefixAsString = newPrefix;
}
var directiveHandlers = {};
function directive(name, callback) {
  directiveHandlers[name] = callback;
  return {
    before(directive2) {
      if (!directiveHandlers[directive2]) {
        console.warn(String.raw`Cannot find directive \`${directive2}\`. \`${name}\` will use the default order of execution`);
        return;
      }
      const pos = directiveOrder.indexOf(directive2);
      directiveOrder.splice(pos >= 0 ? pos : directiveOrder.indexOf("DEFAULT"), 0, name);
    }
  };
}
function directiveExists(name) {
  return Object.keys(directiveHandlers).includes(name);
}
function directives(el, attributes, originalAttributeOverride) {
  attributes = Array.from(attributes);
  if (el._x_virtualDirectives) {
    let vAttributes = Object.entries(el._x_virtualDirectives).map(([name, value]) => ({ name, value }));
    let staticAttributes = attributesOnly(vAttributes);
    vAttributes = vAttributes.map((attribute) => {
      if (staticAttributes.find((attr) => attr.name === attribute.name)) {
        return {
          name: `x-bind:${attribute.name}`,
          value: `"${attribute.value}"`
        };
      }
      return attribute;
    });
    attributes = attributes.concat(vAttributes);
  }
  let transformedAttributeMap = {};
  let directives2 = attributes.map(toTransformedAttributes((newName, oldName) => transformedAttributeMap[newName] = oldName)).filter(outNonAlpineAttributes).map(toParsedDirectives(transformedAttributeMap, originalAttributeOverride)).sort(byPriority);
  return directives2.map((directive2) => {
    return getDirectiveHandler(el, directive2);
  });
}
function attributesOnly(attributes) {
  return Array.from(attributes).map(toTransformedAttributes()).filter((attr) => !outNonAlpineAttributes(attr));
}
var isDeferringHandlers = false;
var directiveHandlerStacks = /* @__PURE__ */ new Map();
var currentHandlerStackKey = Symbol();
function deferHandlingDirectives(callback) {
  isDeferringHandlers = true;
  let key = Symbol();
  currentHandlerStackKey = key;
  directiveHandlerStacks.set(key, []);
  let flushHandlers = () => {
    while (directiveHandlerStacks.get(key).length)
      directiveHandlerStacks.get(key).shift()();
    directiveHandlerStacks.delete(key);
  };
  let stopDeferring = () => {
    isDeferringHandlers = false;
    flushHandlers();
  };
  callback(flushHandlers);
  stopDeferring();
}
function getElementBoundUtilities(el) {
  let cleanups = [];
  let cleanup2 = (callback) => cleanups.push(callback);
  let [effect3, cleanupEffect] = elementBoundEffect(el);
  cleanups.push(cleanupEffect);
  let utilities = {
    Alpine: alpine_default,
    effect: effect3,
    cleanup: cleanup2,
    evaluateLater: evaluateLater.bind(evaluateLater, el),
    evaluate: evaluate.bind(evaluate, el)
  };
  let doCleanup = () => cleanups.forEach((i) => i());
  return [utilities, doCleanup];
}
function getDirectiveHandler(el, directive2) {
  let noop = () => {
  };
  let handler4 = directiveHandlers[directive2.type] || noop;
  let [utilities, cleanup2] = getElementBoundUtilities(el);
  onAttributeRemoved(el, directive2.original, cleanup2);
  let fullHandler = () => {
    if (el._x_ignore || el._x_ignoreSelf)
      return;
    handler4.inline && handler4.inline(el, directive2, utilities);
    handler4 = handler4.bind(handler4, el, directive2, utilities);
    isDeferringHandlers ? directiveHandlerStacks.get(currentHandlerStackKey).push(handler4) : handler4();
  };
  fullHandler.runCleanups = cleanup2;
  return fullHandler;
}
var startingWith = (subject, replacement) => ({ name, value }) => {
  if (name.startsWith(subject))
    name = name.replace(subject, replacement);
  return { name, value };
};
var into = (i) => i;
function toTransformedAttributes(callback = () => {
}) {
  return ({ name, value }) => {
    let { name: newName, value: newValue } = attributeTransformers.reduce((carry, transform) => {
      return transform(carry);
    }, { name, value });
    if (newName !== name)
      callback(newName, name);
    return { name: newName, value: newValue };
  };
}
var attributeTransformers = [];
function mapAttributes(callback) {
  attributeTransformers.push(callback);
}
function outNonAlpineAttributes({ name }) {
  return alpineAttributeRegex().test(name);
}
var alpineAttributeRegex = () => new RegExp(`^${prefixAsString}([^:^.]+)\\b`);
function toParsedDirectives(transformedAttributeMap, originalAttributeOverride) {
  return ({ name, value }) => {
    if (name === value)
      value = "";
    let typeMatch = name.match(alpineAttributeRegex());
    let valueMatch = name.match(/:([a-zA-Z0-9\-_:]+)/);
    let modifiers = name.match(/\.[^.\]]+(?=[^\]]*$)/g) || [];
    let original = originalAttributeOverride || transformedAttributeMap[name] || name;
    return {
      type: typeMatch ? typeMatch[1] : null,
      value: valueMatch ? valueMatch[1] : null,
      modifiers: modifiers.map((i) => i.replace(".", "")),
      expression: value,
      original
    };
  };
}
var DEFAULT = "DEFAULT";
var directiveOrder = [
  "ignore",
  "ref",
  "data",
  "id",
  "anchor",
  "bind",
  "init",
  "for",
  "model",
  "modelable",
  "transition",
  "show",
  "if",
  DEFAULT,
  "teleport"
];
function byPriority(a, b) {
  let typeA = directiveOrder.indexOf(a.type) === -1 ? DEFAULT : a.type;
  let typeB = directiveOrder.indexOf(b.type) === -1 ? DEFAULT : b.type;
  return directiveOrder.indexOf(typeA) - directiveOrder.indexOf(typeB);
}

// packages/alpinejs/src/utils/dispatch.js
function dispatch(el, name, detail = {}) {
  el.dispatchEvent(
    new CustomEvent(name, {
      detail,
      bubbles: true,
      // Allows events to pass the shadow DOM barrier.
      composed: true,
      cancelable: true
    })
  );
}

// packages/alpinejs/src/utils/walk.js
function walk(el, callback) {
  if (typeof ShadowRoot === "function" && el instanceof ShadowRoot) {
    Array.from(el.children).forEach((el2) => walk(el2, callback));
    return;
  }
  let skip = false;
  callback(el, () => skip = true);
  if (skip)
    return;
  let node = el.firstElementChild;
  while (node) {
    walk(node, callback);
    node = node.nextElementSibling;
  }
}

// packages/alpinejs/src/utils/warn.js
function warn(message, ...args) {
  console.warn(`Alpine Warning: ${message}`, ...args);
}

// packages/alpinejs/src/lifecycle.js
var started = false;
function start() {
  if (started)
    warn("Alpine has already been initialized on this page. Calling Alpine.start() more than once can cause problems.");
  started = true;
  if (!document.body)
    warn("Unable to initialize. Trying to load Alpine before `<body>` is available. Did you forget to add `defer` in Alpine's `<script>` tag?");
  dispatch(document, "alpine:init");
  dispatch(document, "alpine:initializing");
  startObservingMutations();
  onElAdded((el) => initTree(el, walk));
  onElRemoved((el) => destroyTree(el));
  onAttributesAdded((el, attrs) => {
    directives(el, attrs).forEach((handle) => handle());
  });
  let outNestedComponents = (el) => !closestRoot(el.parentElement, true);
  Array.from(document.querySelectorAll(allSelectors().join(","))).filter(outNestedComponents).forEach((el) => {
    initTree(el);
  });
  dispatch(document, "alpine:initialized");
  setTimeout(() => {
    warnAboutMissingPlugins();
  });
}
var rootSelectorCallbacks = [];
var initSelectorCallbacks = [];
function rootSelectors() {
  return rootSelectorCallbacks.map((fn) => fn());
}
function allSelectors() {
  return rootSelectorCallbacks.concat(initSelectorCallbacks).map((fn) => fn());
}
function addRootSelector(selectorCallback) {
  rootSelectorCallbacks.push(selectorCallback);
}
function addInitSelector(selectorCallback) {
  initSelectorCallbacks.push(selectorCallback);
}
function closestRoot(el, includeInitSelectors = false) {
  return findClosest(el, (element) => {
    const selectors = includeInitSelectors ? allSelectors() : rootSelectors();
    if (selectors.some((selector) => element.matches(selector)))
      return true;
  });
}
function findClosest(el, callback) {
  if (!el)
    return;
  if (callback(el))
    return el;
  if (el._x_teleportBack)
    el = el._x_teleportBack;
  if (el.parentNode instanceof ShadowRoot) {
    return findClosest(el.parentNode.host, callback);
  }
  if (!el.parentElement)
    return;
  return findClosest(el.parentElement, callback);
}
function isRoot(el) {
  return rootSelectors().some((selector) => el.matches(selector));
}
var initInterceptors2 = [];
function interceptInit(callback) {
  initInterceptors2.push(callback);
}
var markerDispenser = 1;
function initTree(el, walker = walk, intercept = () => {
}) {
  if (findClosest(el, (i) => i._x_ignore))
    return;
  deferHandlingDirectives(() => {
    walker(el, (el2, skip) => {
      if (el2._x_marker)
        return;
      intercept(el2, skip);
      initInterceptors2.forEach((i) => i(el2, skip));
      directives(el2, el2.attributes).forEach((handle) => handle());
      if (!el2._x_ignore)
        el2._x_marker = markerDispenser++;
      el2._x_ignore && skip();
    });
  });
}
function destroyTree(root, walker = walk) {
  walker(root, (el) => {
    cleanupElement(el);
    cleanupAttributes(el);
    delete el._x_marker;
  });
}
function warnAboutMissingPlugins() {
  let pluginDirectives = [
    ["ui", "dialog", ["[x-dialog], [x-popover]"]],
    ["anchor", "anchor", ["[x-anchor]"]],
    ["sort", "sort", ["[x-sort]"]]
  ];
  pluginDirectives.forEach(([plugin2, directive2, selectors]) => {
    if (directiveExists(directive2))
      return;
    selectors.some((selector) => {
      if (document.querySelector(selector)) {
        warn(`found "${selector}", but missing ${plugin2} plugin`);
        return true;
      }
    });
  });
}

// packages/alpinejs/src/nextTick.js
var tickStack = [];
var isHolding = false;
function nextTick(callback = () => {
}) {
  queueMicrotask(() => {
    isHolding || setTimeout(() => {
      releaseNextTicks();
    });
  });
  return new Promise((res) => {
    tickStack.push(() => {
      callback();
      res();
    });
  });
}
function releaseNextTicks() {
  isHolding = false;
  while (tickStack.length)
    tickStack.shift()();
}
function holdNextTicks() {
  isHolding = true;
}

// packages/alpinejs/src/utils/classes.js
function setClasses(el, value) {
  if (Array.isArray(value)) {
    return setClassesFromString(el, value.join(" "));
  } else if (typeof value === "object" && value !== null) {
    return setClassesFromObject(el, value);
  } else if (typeof value === "function") {
    return setClasses(el, value());
  }
  return setClassesFromString(el, value);
}
function setClassesFromString(el, classString) {
  let missingClasses = (classString2) => classString2.split(" ").filter((i) => !el.classList.contains(i)).filter(Boolean);
  let addClassesAndReturnUndo = (classes) => {
    el.classList.add(...classes);
    return () => {
      el.classList.remove(...classes);
    };
  };
  classString = classString === true ? classString = "" : classString || "";
  return addClassesAndReturnUndo(missingClasses(classString));
}
function setClassesFromObject(el, classObject) {
  let split = (classString) => classString.split(" ").filter(Boolean);
  let forAdd = Object.entries(classObject).flatMap(([classString, bool]) => bool ? split(classString) : false).filter(Boolean);
  let forRemove = Object.entries(classObject).flatMap(([classString, bool]) => !bool ? split(classString) : false).filter(Boolean);
  let added = [];
  let removed = [];
  forRemove.forEach((i) => {
    if (el.classList.contains(i)) {
      el.classList.remove(i);
      removed.push(i);
    }
  });
  forAdd.forEach((i) => {
    if (!el.classList.contains(i)) {
      el.classList.add(i);
      added.push(i);
    }
  });
  return () => {
    removed.forEach((i) => el.classList.add(i));
    added.forEach((i) => el.classList.remove(i));
  };
}

// packages/alpinejs/src/utils/styles.js
function setStyles(el, value) {
  if (typeof value === "object" && value !== null) {
    return setStylesFromObject(el, value);
  }
  return setStylesFromString(el, value);
}
function setStylesFromObject(el, value) {
  let previousStyles = {};
  Object.entries(value).forEach(([key, value2]) => {
    previousStyles[key] = el.style[key];
    if (!key.startsWith("--")) {
      key = kebabCase(key);
    }
    el.style.setProperty(key, value2);
  });
  setTimeout(() => {
    if (el.style.length === 0) {
      el.removeAttribute("style");
    }
  });
  return () => {
    setStyles(el, previousStyles);
  };
}
function setStylesFromString(el, value) {
  let cache = el.getAttribute("style", value);
  el.setAttribute("style", value);
  return () => {
    el.setAttribute("style", cache || "");
  };
}
function kebabCase(subject) {
  return subject.replace(/([a-z])([A-Z])/g, "$1-$2").toLowerCase();
}

// packages/alpinejs/src/utils/once.js
function once(callback, fallback = () => {
}) {
  let called = false;
  return function() {
    if (!called) {
      called = true;
      callback.apply(this, arguments);
    } else {
      fallback.apply(this, arguments);
    }
  };
}

// packages/alpinejs/src/directives/x-transition.js
directive("transition", (el, { value, modifiers, expression }, { evaluate: evaluate2 }) => {
  if (typeof expression === "function")
    expression = evaluate2(expression);
  if (expression === false)
    return;
  if (!expression || typeof expression === "boolean") {
    registerTransitionsFromHelper(el, modifiers, value);
  } else {
    registerTransitionsFromClassString(el, expression, value);
  }
});
function registerTransitionsFromClassString(el, classString, stage) {
  registerTransitionObject(el, setClasses, "");
  let directiveStorageMap = {
    "enter": (classes) => {
      el._x_transition.enter.during = classes;
    },
    "enter-start": (classes) => {
      el._x_transition.enter.start = classes;
    },
    "enter-end": (classes) => {
      el._x_transition.enter.end = classes;
    },
    "leave": (classes) => {
      el._x_transition.leave.during = classes;
    },
    "leave-start": (classes) => {
      el._x_transition.leave.start = classes;
    },
    "leave-end": (classes) => {
      el._x_transition.leave.end = classes;
    }
  };
  directiveStorageMap[stage](classString);
}
function registerTransitionsFromHelper(el, modifiers, stage) {
  registerTransitionObject(el, setStyles);
  let doesntSpecify = !modifiers.includes("in") && !modifiers.includes("out") && !stage;
  let transitioningIn = doesntSpecify || modifiers.includes("in") || ["enter"].includes(stage);
  let transitioningOut = doesntSpecify || modifiers.includes("out") || ["leave"].includes(stage);
  if (modifiers.includes("in") && !doesntSpecify) {
    modifiers = modifiers.filter((i, index) => index < modifiers.indexOf("out"));
  }
  if (modifiers.includes("out") && !doesntSpecify) {
    modifiers = modifiers.filter((i, index) => index > modifiers.indexOf("out"));
  }
  let wantsAll = !modifiers.includes("opacity") && !modifiers.includes("scale");
  let wantsOpacity = wantsAll || modifiers.includes("opacity");
  let wantsScale = wantsAll || modifiers.includes("scale");
  let opacityValue = wantsOpacity ? 0 : 1;
  let scaleValue = wantsScale ? modifierValue(modifiers, "scale", 95) / 100 : 1;
  let delay = modifierValue(modifiers, "delay", 0) / 1e3;
  let origin = modifierValue(modifiers, "origin", "center");
  let property = "opacity, transform";
  let durationIn = modifierValue(modifiers, "duration", 150) / 1e3;
  let durationOut = modifierValue(modifiers, "duration", 75) / 1e3;
  let easing = `cubic-bezier(0.4, 0.0, 0.2, 1)`;
  if (transitioningIn) {
    el._x_transition.enter.during = {
      transformOrigin: origin,
      transitionDelay: `${delay}s`,
      transitionProperty: property,
      transitionDuration: `${durationIn}s`,
      transitionTimingFunction: easing
    };
    el._x_transition.enter.start = {
      opacity: opacityValue,
      transform: `scale(${scaleValue})`
    };
    el._x_transition.enter.end = {
      opacity: 1,
      transform: `scale(1)`
    };
  }
  if (transitioningOut) {
    el._x_transition.leave.during = {
      transformOrigin: origin,
      transitionDelay: `${delay}s`,
      transitionProperty: property,
      transitionDuration: `${durationOut}s`,
      transitionTimingFunction: easing
    };
    el._x_transition.leave.start = {
      opacity: 1,
      transform: `scale(1)`
    };
    el._x_transition.leave.end = {
      opacity: opacityValue,
      transform: `scale(${scaleValue})`
    };
  }
}
function registerTransitionObject(el, setFunction, defaultValue = {}) {
  if (!el._x_transition)
    el._x_transition = {
      enter: { during: defaultValue, start: defaultValue, end: defaultValue },
      leave: { during: defaultValue, start: defaultValue, end: defaultValue },
      in(before = () => {
      }, after = () => {
      }) {
        transition(el, setFunction, {
          during: this.enter.during,
          start: this.enter.start,
          end: this.enter.end
        }, before, after);
      },
      out(before = () => {
      }, after = () => {
      }) {
        transition(el, setFunction, {
          during: this.leave.during,
          start: this.leave.start,
          end: this.leave.end
        }, before, after);
      }
    };
}
window.Element.prototype._x_toggleAndCascadeWithTransitions = function(el, value, show, hide) {
  const nextTick2 = document.visibilityState === "visible" ? requestAnimationFrame : setTimeout;
  let clickAwayCompatibleShow = () => nextTick2(show);
  if (value) {
    if (el._x_transition && (el._x_transition.enter || el._x_transition.leave)) {
      el._x_transition.enter && (Object.entries(el._x_transition.enter.during).length || Object.entries(el._x_transition.enter.start).length || Object.entries(el._x_transition.enter.end).length) ? el._x_transition.in(show) : clickAwayCompatibleShow();
    } else {
      el._x_transition ? el._x_transition.in(show) : clickAwayCompatibleShow();
    }
    return;
  }
  el._x_hidePromise = el._x_transition ? new Promise((resolve, reject) => {
    el._x_transition.out(() => {
    }, () => resolve(hide));
    el._x_transitioning && el._x_transitioning.beforeCancel(() => reject({ isFromCancelledTransition: true }));
  }) : Promise.resolve(hide);
  queueMicrotask(() => {
    let closest = closestHide(el);
    if (closest) {
      if (!closest._x_hideChildren)
        closest._x_hideChildren = [];
      closest._x_hideChildren.push(el);
    } else {
      nextTick2(() => {
        let hideAfterChildren = (el2) => {
          let carry = Promise.all([
            el2._x_hidePromise,
            ...(el2._x_hideChildren || []).map(hideAfterChildren)
          ]).then(([i]) => i?.());
          delete el2._x_hidePromise;
          delete el2._x_hideChildren;
          return carry;
        };
        hideAfterChildren(el).catch((e) => {
          if (!e.isFromCancelledTransition)
            throw e;
        });
      });
    }
  });
};
function closestHide(el) {
  let parent = el.parentNode;
  if (!parent)
    return;
  return parent._x_hidePromise ? parent : closestHide(parent);
}
function transition(el, setFunction, { during, start: start2, end } = {}, before = () => {
}, after = () => {
}) {
  if (el._x_transitioning)
    el._x_transitioning.cancel();
  if (Object.keys(during).length === 0 && Object.keys(start2).length === 0 && Object.keys(end).length === 0) {
    before();
    after();
    return;
  }
  let undoStart, undoDuring, undoEnd;
  performTransition(el, {
    start() {
      undoStart = setFunction(el, start2);
    },
    during() {
      undoDuring = setFunction(el, during);
    },
    before,
    end() {
      undoStart();
      undoEnd = setFunction(el, end);
    },
    after,
    cleanup() {
      undoDuring();
      undoEnd();
    }
  });
}
function performTransition(el, stages) {
  let interrupted, reachedBefore, reachedEnd;
  let finish = once(() => {
    mutateDom(() => {
      interrupted = true;
      if (!reachedBefore)
        stages.before();
      if (!reachedEnd) {
        stages.end();
        releaseNextTicks();
      }
      stages.after();
      if (el.isConnected)
        stages.cleanup();
      delete el._x_transitioning;
    });
  });
  el._x_transitioning = {
    beforeCancels: [],
    beforeCancel(callback) {
      this.beforeCancels.push(callback);
    },
    cancel: once(function() {
      while (this.beforeCancels.length) {
        this.beforeCancels.shift()();
      }
      finish();
    }),
    finish
  };
  mutateDom(() => {
    stages.start();
    stages.during();
  });
  holdNextTicks();
  requestAnimationFrame(() => {
    if (interrupted)
      return;
    let duration = Number(getComputedStyle(el).transitionDuration.replace(/,.*/, "").replace("s", "")) * 1e3;
    let delay = Number(getComputedStyle(el).transitionDelay.replace(/,.*/, "").replace("s", "")) * 1e3;
    if (duration === 0)
      duration = Number(getComputedStyle(el).animationDuration.replace("s", "")) * 1e3;
    mutateDom(() => {
      stages.before();
    });
    reachedBefore = true;
    requestAnimationFrame(() => {
      if (interrupted)
        return;
      mutateDom(() => {
        stages.end();
      });
      releaseNextTicks();
      setTimeout(el._x_transitioning.finish, duration + delay);
      reachedEnd = true;
    });
  });
}
function modifierValue(modifiers, key, fallback) {
  if (modifiers.indexOf(key) === -1)
    return fallback;
  const rawValue = modifiers[modifiers.indexOf(key) + 1];
  if (!rawValue)
    return fallback;
  if (key === "scale") {
    if (isNaN(rawValue))
      return fallback;
  }
  if (key === "duration" || key === "delay") {
    let match = rawValue.match(/([0-9]+)ms/);
    if (match)
      return match[1];
  }
  if (key === "origin") {
    if (["top", "right", "left", "center", "bottom"].includes(modifiers[modifiers.indexOf(key) + 2])) {
      return [rawValue, modifiers[modifiers.indexOf(key) + 2]].join(" ");
    }
  }
  return rawValue;
}

// packages/alpinejs/src/clone.js
var isCloning = false;
function skipDuringClone(callback, fallback = () => {
}) {
  return (...args) => isCloning ? fallback(...args) : callback(...args);
}
function onlyDuringClone(callback) {
  return (...args) => isCloning && callback(...args);
}
var interceptors = [];
function interceptClone(callback) {
  interceptors.push(callback);
}
function cloneNode(from, to) {
  interceptors.forEach((i) => i(from, to));
  isCloning = true;
  dontRegisterReactiveSideEffects(() => {
    initTree(to, (el, callback) => {
      callback(el, () => {
      });
    });
  });
  isCloning = false;
}
var isCloningLegacy = false;
function clone(oldEl, newEl) {
  if (!newEl._x_dataStack)
    newEl._x_dataStack = oldEl._x_dataStack;
  isCloning = true;
  isCloningLegacy = true;
  dontRegisterReactiveSideEffects(() => {
    cloneTree(newEl);
  });
  isCloning = false;
  isCloningLegacy = false;
}
function cloneTree(el) {
  let hasRunThroughFirstEl = false;
  let shallowWalker = (el2, callback) => {
    walk(el2, (el3, skip) => {
      if (hasRunThroughFirstEl && isRoot(el3))
        return skip();
      hasRunThroughFirstEl = true;
      callback(el3, skip);
    });
  };
  initTree(el, shallowWalker);
}
function dontRegisterReactiveSideEffects(callback) {
  let cache = effect;
  overrideEffect((callback2, el) => {
    let storedEffect = cache(callback2);
    release(storedEffect);
    return () => {
    };
  });
  callback();
  overrideEffect(cache);
}

// packages/alpinejs/src/utils/bind.js
function bind(el, name, value, modifiers = []) {
  if (!el._x_bindings)
    el._x_bindings = reactive({});
  el._x_bindings[name] = value;
  name = modifiers.includes("camel") ? camelCase(name) : name;
  switch (name) {
    case "value":
      bindInputValue(el, value);
      break;
    case "style":
      bindStyles(el, value);
      break;
    case "class":
      bindClasses(el, value);
      break;
    case "selected":
    case "checked":
      bindAttributeAndProperty(el, name, value);
      break;
    default:
      bindAttribute(el, name, value);
      break;
  }
}
function bindInputValue(el, value) {
  if (isRadio(el)) {
    if (el.attributes.value === void 0) {
      el.value = value;
    }
    if (window.fromModel) {
      if (typeof value === "boolean") {
        el.checked = safeParseBoolean(el.value) === value;
      } else {
        el.checked = checkedAttrLooseCompare(el.value, value);
      }
    }
  } else if (isCheckbox(el)) {
    if (Number.isInteger(value)) {
      el.value = value;
    } else if (!Array.isArray(value) && typeof value !== "boolean" && ![null, void 0].includes(value)) {
      el.value = String(value);
    } else {
      if (Array.isArray(value)) {
        el.checked = value.some((val) => checkedAttrLooseCompare(val, el.value));
      } else {
        el.checked = !!value;
      }
    }
  } else if (el.tagName === "SELECT") {
    updateSelect(el, value);
  } else {
    if (el.value === value)
      return;
    el.value = value === void 0 ? "" : value;
  }
}
function bindClasses(el, value) {
  if (el._x_undoAddedClasses)
    el._x_undoAddedClasses();
  el._x_undoAddedClasses = setClasses(el, value);
}
function bindStyles(el, value) {
  if (el._x_undoAddedStyles)
    el._x_undoAddedStyles();
  el._x_undoAddedStyles = setStyles(el, value);
}
function bindAttributeAndProperty(el, name, value) {
  bindAttribute(el, name, value);
  setPropertyIfChanged(el, name, value);
}
function bindAttribute(el, name, value) {
  if ([null, void 0, false].includes(value) && attributeShouldntBePreservedIfFalsy(name)) {
    el.removeAttribute(name);
  } else {
    if (isBooleanAttr(name))
      value = name;
    setIfChanged(el, name, value);
  }
}
function setIfChanged(el, attrName, value) {
  if (el.getAttribute(attrName) != value) {
    el.setAttribute(attrName, value);
  }
}
function setPropertyIfChanged(el, propName, value) {
  if (el[propName] !== value) {
    el[propName] = value;
  }
}
function updateSelect(el, value) {
  const arrayWrappedValue = [].concat(value).map((value2) => {
    return value2 + "";
  });
  Array.from(el.options).forEach((option) => {
    option.selected = arrayWrappedValue.includes(option.value);
  });
}
function camelCase(subject) {
  return subject.toLowerCase().replace(/-(\w)/g, (match, char) => char.toUpperCase());
}
function checkedAttrLooseCompare(valueA, valueB) {
  return valueA == valueB;
}
function safeParseBoolean(rawValue) {
  if ([1, "1", "true", "on", "yes", true].includes(rawValue)) {
    return true;
  }
  if ([0, "0", "false", "off", "no", false].includes(rawValue)) {
    return false;
  }
  return rawValue ? Boolean(rawValue) : null;
}
var booleanAttributes = /* @__PURE__ */ new Set([
  "allowfullscreen",
  "async",
  "autofocus",
  "autoplay",
  "checked",
  "controls",
  "default",
  "defer",
  "disabled",
  "formnovalidate",
  "inert",
  "ismap",
  "itemscope",
  "loop",
  "multiple",
  "muted",
  "nomodule",
  "novalidate",
  "open",
  "playsinline",
  "readonly",
  "required",
  "reversed",
  "selected",
  "shadowrootclonable",
  "shadowrootdelegatesfocus",
  "shadowrootserializable"
]);
function isBooleanAttr(attrName) {
  return booleanAttributes.has(attrName);
}
function attributeShouldntBePreservedIfFalsy(name) {
  return !["aria-pressed", "aria-checked", "aria-expanded", "aria-selected"].includes(name);
}
function getBinding(el, name, fallback) {
  if (el._x_bindings && el._x_bindings[name] !== void 0)
    return el._x_bindings[name];
  return getAttributeBinding(el, name, fallback);
}
function extractProp(el, name, fallback, extract = true) {
  if (el._x_bindings && el._x_bindings[name] !== void 0)
    return el._x_bindings[name];
  if (el._x_inlineBindings && el._x_inlineBindings[name] !== void 0) {
    let binding = el._x_inlineBindings[name];
    binding.extract = extract;
    return dontAutoEvaluateFunctions(() => {
      return evaluate(el, binding.expression);
    });
  }
  return getAttributeBinding(el, name, fallback);
}
function getAttributeBinding(el, name, fallback) {
  let attr = el.getAttribute(name);
  if (attr === null)
    return typeof fallback === "function" ? fallback() : fallback;
  if (attr === "")
    return true;
  if (isBooleanAttr(name)) {
    return !![name, "true"].includes(attr);
  }
  return attr;
}
function isCheckbox(el) {
  return el.type === "checkbox" || el.localName === "ui-checkbox" || el.localName === "ui-switch";
}
function isRadio(el) {
  return el.type === "radio" || el.localName === "ui-radio";
}

// packages/alpinejs/src/utils/debounce.js
function debounce(func, wait) {
  let timeout;
  return function() {
    const context = this, args = arguments;
    const later = function() {
      timeout = null;
      func.apply(context, args);
    };
    clearTimeout(timeout);
    timeout = setTimeout(later, wait);
  };
}

// packages/alpinejs/src/utils/throttle.js
function throttle(func, limit) {
  let inThrottle;
  return function() {
    let context = this, args = arguments;
    if (!inThrottle) {
      func.apply(context, args);
      inThrottle = true;
      setTimeout(() => inThrottle = false, limit);
    }
  };
}

// packages/alpinejs/src/entangle.js
function entangle({ get: outerGet, set: outerSet }, { get: innerGet, set: innerSet }) {
  let firstRun = true;
  let outerHash;
  let reference = effect(() => {
    let outer = outerGet();
    let inner = innerGet();
    if (firstRun) {
      innerSet(cloneIfObject(outer));
      firstRun = false;
    } else {
      let outerHashLatest = JSON.stringify(outer);
      let innerHashLatest = JSON.stringify(inner);
      if (outerHashLatest !== outerHash) {
        innerSet(cloneIfObject(outer));
      } else if (outerHashLatest !== innerHashLatest) {
        outerSet(cloneIfObject(inner));
      } else ;
    }
    outerHash = JSON.stringify(outerGet());
    JSON.stringify(innerGet());
  });
  return () => {
    release(reference);
  };
}
function cloneIfObject(value) {
  return typeof value === "object" ? JSON.parse(JSON.stringify(value)) : value;
}

// packages/alpinejs/src/plugin.js
function plugin(callback) {
  let callbacks = Array.isArray(callback) ? callback : [callback];
  callbacks.forEach((i) => i(alpine_default));
}

// packages/alpinejs/src/store.js
var stores = {};
var isReactive = false;
function store(name, value) {
  if (!isReactive) {
    stores = reactive(stores);
    isReactive = true;
  }
  if (value === void 0) {
    return stores[name];
  }
  stores[name] = value;
  initInterceptors(stores[name]);
  if (typeof value === "object" && value !== null && value.hasOwnProperty("init") && typeof value.init === "function") {
    stores[name].init();
  }
}
function getStores() {
  return stores;
}

// packages/alpinejs/src/binds.js
var binds = {};
function bind2(name, bindings) {
  let getBindings = typeof bindings !== "function" ? () => bindings : bindings;
  if (name instanceof Element) {
    return applyBindingsObject(name, getBindings());
  } else {
    binds[name] = getBindings;
  }
  return () => {
  };
}
function injectBindingProviders(obj) {
  Object.entries(binds).forEach(([name, callback]) => {
    Object.defineProperty(obj, name, {
      get() {
        return (...args) => {
          return callback(...args);
        };
      }
    });
  });
  return obj;
}
function applyBindingsObject(el, obj, original) {
  let cleanupRunners = [];
  while (cleanupRunners.length)
    cleanupRunners.pop()();
  let attributes = Object.entries(obj).map(([name, value]) => ({ name, value }));
  let staticAttributes = attributesOnly(attributes);
  attributes = attributes.map((attribute) => {
    if (staticAttributes.find((attr) => attr.name === attribute.name)) {
      return {
        name: `x-bind:${attribute.name}`,
        value: `"${attribute.value}"`
      };
    }
    return attribute;
  });
  directives(el, attributes, original).map((handle) => {
    cleanupRunners.push(handle.runCleanups);
    handle();
  });
  return () => {
    while (cleanupRunners.length)
      cleanupRunners.pop()();
  };
}

// packages/alpinejs/src/datas.js
var datas = {};
function data(name, callback) {
  datas[name] = callback;
}
function injectDataProviders(obj, context) {
  Object.entries(datas).forEach(([name, callback]) => {
    Object.defineProperty(obj, name, {
      get() {
        return (...args) => {
          return callback.bind(context)(...args);
        };
      },
      enumerable: false
    });
  });
  return obj;
}

// packages/alpinejs/src/alpine.js
var Alpine = {
  get reactive() {
    return reactive;
  },
  get release() {
    return release;
  },
  get effect() {
    return effect;
  },
  get raw() {
    return raw;
  },
  version: "3.15.4",
  flushAndStopDeferringMutations,
  dontAutoEvaluateFunctions,
  disableEffectScheduling,
  startObservingMutations,
  stopObservingMutations,
  setReactivityEngine,
  onAttributeRemoved,
  onAttributesAdded,
  closestDataStack,
  skipDuringClone,
  onlyDuringClone,
  addRootSelector,
  addInitSelector,
  setErrorHandler,
  interceptClone,
  addScopeToNode,
  deferMutations,
  mapAttributes,
  evaluateLater,
  interceptInit,
  initInterceptors,
  injectMagics,
  setEvaluator,
  setRawEvaluator,
  mergeProxies,
  extractProp,
  findClosest,
  onElRemoved,
  closestRoot,
  destroyTree,
  interceptor,
  // INTERNAL: not public API and is subject to change without major release.
  transition,
  // INTERNAL
  setStyles,
  // INTERNAL
  mutateDom,
  directive,
  entangle,
  throttle,
  debounce,
  evaluate,
  evaluateRaw,
  initTree,
  nextTick,
  prefixed: prefix,
  prefix: setPrefix,
  plugin,
  magic,
  store,
  start,
  clone,
  // INTERNAL
  cloneNode,
  // INTERNAL
  bound: getBinding,
  $data: scope,
  watch,
  walk,
  data,
  bind: bind2
};
var alpine_default = Alpine;

// node_modules/@vue/shared/dist/shared.esm-bundler.js
function makeMap(str, expectsLowerCase) {
  const map = /* @__PURE__ */ Object.create(null);
  const list = str.split(",");
  for (let i = 0; i < list.length; i++) {
    map[list[i]] = true;
  }
  return (val) => !!map[val];
}
var EMPTY_OBJ = Object.freeze({}) ;
var hasOwnProperty = Object.prototype.hasOwnProperty;
var hasOwn = (val, key) => hasOwnProperty.call(val, key);
var isArray = Array.isArray;
var isMap = (val) => toTypeString(val) === "[object Map]";
var isString = (val) => typeof val === "string";
var isSymbol = (val) => typeof val === "symbol";
var isObject = (val) => val !== null && typeof val === "object";
var objectToString = Object.prototype.toString;
var toTypeString = (value) => objectToString.call(value);
var toRawType = (value) => {
  return toTypeString(value).slice(8, -1);
};
var isIntegerKey = (key) => isString(key) && key !== "NaN" && key[0] !== "-" && "" + parseInt(key, 10) === key;
var cacheStringFunction = (fn) => {
  const cache = /* @__PURE__ */ Object.create(null);
  return (str) => {
    const hit = cache[str];
    return hit || (cache[str] = fn(str));
  };
};
var capitalize = cacheStringFunction((str) => str.charAt(0).toUpperCase() + str.slice(1));
var hasChanged = (value, oldValue) => value !== oldValue && (value === value || oldValue === oldValue);

// node_modules/@vue/reactivity/dist/reactivity.esm-bundler.js
var targetMap = /* @__PURE__ */ new WeakMap();
var effectStack = [];
var activeEffect;
var ITERATE_KEY = Symbol("iterate" );
var MAP_KEY_ITERATE_KEY = Symbol("Map key iterate" );
function isEffect(fn) {
  return fn && fn._isEffect === true;
}
function effect2(fn, options = EMPTY_OBJ) {
  if (isEffect(fn)) {
    fn = fn.raw;
  }
  const effect3 = createReactiveEffect(fn, options);
  if (!options.lazy) {
    effect3();
  }
  return effect3;
}
function stop(effect3) {
  if (effect3.active) {
    cleanup(effect3);
    if (effect3.options.onStop) {
      effect3.options.onStop();
    }
    effect3.active = false;
  }
}
var uid = 0;
function createReactiveEffect(fn, options) {
  const effect3 = function reactiveEffect() {
    if (!effect3.active) {
      return fn();
    }
    if (!effectStack.includes(effect3)) {
      cleanup(effect3);
      try {
        enableTracking();
        effectStack.push(effect3);
        activeEffect = effect3;
        return fn();
      } finally {
        effectStack.pop();
        resetTracking();
        activeEffect = effectStack[effectStack.length - 1];
      }
    }
  };
  effect3.id = uid++;
  effect3.allowRecurse = !!options.allowRecurse;
  effect3._isEffect = true;
  effect3.active = true;
  effect3.raw = fn;
  effect3.deps = [];
  effect3.options = options;
  return effect3;
}
function cleanup(effect3) {
  const { deps } = effect3;
  if (deps.length) {
    for (let i = 0; i < deps.length; i++) {
      deps[i].delete(effect3);
    }
    deps.length = 0;
  }
}
var shouldTrack = true;
var trackStack = [];
function pauseTracking() {
  trackStack.push(shouldTrack);
  shouldTrack = false;
}
function enableTracking() {
  trackStack.push(shouldTrack);
  shouldTrack = true;
}
function resetTracking() {
  const last = trackStack.pop();
  shouldTrack = last === void 0 ? true : last;
}
function track(target, type, key) {
  if (!shouldTrack || activeEffect === void 0) {
    return;
  }
  let depsMap = targetMap.get(target);
  if (!depsMap) {
    targetMap.set(target, depsMap = /* @__PURE__ */ new Map());
  }
  let dep = depsMap.get(key);
  if (!dep) {
    depsMap.set(key, dep = /* @__PURE__ */ new Set());
  }
  if (!dep.has(activeEffect)) {
    dep.add(activeEffect);
    activeEffect.deps.push(dep);
    if (activeEffect.options.onTrack) {
      activeEffect.options.onTrack({
        effect: activeEffect,
        target,
        type,
        key
      });
    }
  }
}
function trigger(target, type, key, newValue, oldValue, oldTarget) {
  const depsMap = targetMap.get(target);
  if (!depsMap) {
    return;
  }
  const effects = /* @__PURE__ */ new Set();
  const add2 = (effectsToAdd) => {
    if (effectsToAdd) {
      effectsToAdd.forEach((effect3) => {
        if (effect3 !== activeEffect || effect3.allowRecurse) {
          effects.add(effect3);
        }
      });
    }
  };
  if (type === "clear") {
    depsMap.forEach(add2);
  } else if (key === "length" && isArray(target)) {
    depsMap.forEach((dep, key2) => {
      if (key2 === "length" || key2 >= newValue) {
        add2(dep);
      }
    });
  } else {
    if (key !== void 0) {
      add2(depsMap.get(key));
    }
    switch (type) {
      case "add":
        if (!isArray(target)) {
          add2(depsMap.get(ITERATE_KEY));
          if (isMap(target)) {
            add2(depsMap.get(MAP_KEY_ITERATE_KEY));
          }
        } else if (isIntegerKey(key)) {
          add2(depsMap.get("length"));
        }
        break;
      case "delete":
        if (!isArray(target)) {
          add2(depsMap.get(ITERATE_KEY));
          if (isMap(target)) {
            add2(depsMap.get(MAP_KEY_ITERATE_KEY));
          }
        }
        break;
      case "set":
        if (isMap(target)) {
          add2(depsMap.get(ITERATE_KEY));
        }
        break;
    }
  }
  const run = (effect3) => {
    if (effect3.options.onTrigger) {
      effect3.options.onTrigger({
        effect: effect3,
        target,
        key,
        type,
        newValue,
        oldValue,
        oldTarget
      });
    }
    if (effect3.options.scheduler) {
      effect3.options.scheduler(effect3);
    } else {
      effect3();
    }
  };
  effects.forEach(run);
}
var isNonTrackableKeys = /* @__PURE__ */ makeMap(`__proto__,__v_isRef,__isVue`);
var builtInSymbols = new Set(Object.getOwnPropertyNames(Symbol).map((key) => Symbol[key]).filter(isSymbol));
var get2 = /* @__PURE__ */ createGetter();
var readonlyGet = /* @__PURE__ */ createGetter(true);
var arrayInstrumentations = /* @__PURE__ */ createArrayInstrumentations();
function createArrayInstrumentations() {
  const instrumentations = {};
  ["includes", "indexOf", "lastIndexOf"].forEach((key) => {
    instrumentations[key] = function(...args) {
      const arr = toRaw(this);
      for (let i = 0, l = this.length; i < l; i++) {
        track(arr, "get", i + "");
      }
      const res = arr[key](...args);
      if (res === -1 || res === false) {
        return arr[key](...args.map(toRaw));
      } else {
        return res;
      }
    };
  });
  ["push", "pop", "shift", "unshift", "splice"].forEach((key) => {
    instrumentations[key] = function(...args) {
      pauseTracking();
      const res = toRaw(this)[key].apply(this, args);
      resetTracking();
      return res;
    };
  });
  return instrumentations;
}
function createGetter(isReadonly = false, shallow = false) {
  return function get3(target, key, receiver) {
    if (key === "__v_isReactive") {
      return !isReadonly;
    } else if (key === "__v_isReadonly") {
      return isReadonly;
    } else if (key === "__v_raw" && receiver === (isReadonly ? shallow ? shallowReadonlyMap : readonlyMap : shallow ? shallowReactiveMap : reactiveMap).get(target)) {
      return target;
    }
    const targetIsArray = isArray(target);
    if (!isReadonly && targetIsArray && hasOwn(arrayInstrumentations, key)) {
      return Reflect.get(arrayInstrumentations, key, receiver);
    }
    const res = Reflect.get(target, key, receiver);
    if (isSymbol(key) ? builtInSymbols.has(key) : isNonTrackableKeys(key)) {
      return res;
    }
    if (!isReadonly) {
      track(target, "get", key);
    }
    if (shallow) {
      return res;
    }
    if (isRef(res)) {
      const shouldUnwrap = !targetIsArray || !isIntegerKey(key);
      return shouldUnwrap ? res.value : res;
    }
    if (isObject(res)) {
      return isReadonly ? readonly(res) : reactive2(res);
    }
    return res;
  };
}
var set2 = /* @__PURE__ */ createSetter();
function createSetter(shallow = false) {
  return function set3(target, key, value, receiver) {
    let oldValue = target[key];
    if (!shallow) {
      value = toRaw(value);
      oldValue = toRaw(oldValue);
      if (!isArray(target) && isRef(oldValue) && !isRef(value)) {
        oldValue.value = value;
        return true;
      }
    }
    const hadKey = isArray(target) && isIntegerKey(key) ? Number(key) < target.length : hasOwn(target, key);
    const result = Reflect.set(target, key, value, receiver);
    if (target === toRaw(receiver)) {
      if (!hadKey) {
        trigger(target, "add", key, value);
      } else if (hasChanged(value, oldValue)) {
        trigger(target, "set", key, value, oldValue);
      }
    }
    return result;
  };
}
function deleteProperty(target, key) {
  const hadKey = hasOwn(target, key);
  const oldValue = target[key];
  const result = Reflect.deleteProperty(target, key);
  if (result && hadKey) {
    trigger(target, "delete", key, void 0, oldValue);
  }
  return result;
}
function has(target, key) {
  const result = Reflect.has(target, key);
  if (!isSymbol(key) || !builtInSymbols.has(key)) {
    track(target, "has", key);
  }
  return result;
}
function ownKeys(target) {
  track(target, "iterate", isArray(target) ? "length" : ITERATE_KEY);
  return Reflect.ownKeys(target);
}
var mutableHandlers = {
  get: get2,
  set: set2,
  deleteProperty,
  has,
  ownKeys
};
var readonlyHandlers = {
  get: readonlyGet,
  set(target, key) {
    {
      console.warn(`Set operation on key "${String(key)}" failed: target is readonly.`, target);
    }
    return true;
  },
  deleteProperty(target, key) {
    {
      console.warn(`Delete operation on key "${String(key)}" failed: target is readonly.`, target);
    }
    return true;
  }
};
var toReactive = (value) => isObject(value) ? reactive2(value) : value;
var toReadonly = (value) => isObject(value) ? readonly(value) : value;
var toShallow = (value) => value;
var getProto = (v) => Reflect.getPrototypeOf(v);
function get$1(target, key, isReadonly = false, isShallow = false) {
  target = target[
    "__v_raw"
    /* RAW */
  ];
  const rawTarget = toRaw(target);
  const rawKey = toRaw(key);
  if (key !== rawKey) {
    !isReadonly && track(rawTarget, "get", key);
  }
  !isReadonly && track(rawTarget, "get", rawKey);
  const { has: has2 } = getProto(rawTarget);
  const wrap = isShallow ? toShallow : isReadonly ? toReadonly : toReactive;
  if (has2.call(rawTarget, key)) {
    return wrap(target.get(key));
  } else if (has2.call(rawTarget, rawKey)) {
    return wrap(target.get(rawKey));
  } else if (target !== rawTarget) {
    target.get(key);
  }
}
function has$1(key, isReadonly = false) {
  const target = this[
    "__v_raw"
    /* RAW */
  ];
  const rawTarget = toRaw(target);
  const rawKey = toRaw(key);
  if (key !== rawKey) {
    !isReadonly && track(rawTarget, "has", key);
  }
  !isReadonly && track(rawTarget, "has", rawKey);
  return key === rawKey ? target.has(key) : target.has(key) || target.has(rawKey);
}
function size(target, isReadonly = false) {
  target = target[
    "__v_raw"
    /* RAW */
  ];
  !isReadonly && track(toRaw(target), "iterate", ITERATE_KEY);
  return Reflect.get(target, "size", target);
}
function add(value) {
  value = toRaw(value);
  const target = toRaw(this);
  const proto = getProto(target);
  const hadKey = proto.has.call(target, value);
  if (!hadKey) {
    target.add(value);
    trigger(target, "add", value, value);
  }
  return this;
}
function set$1(key, value) {
  value = toRaw(value);
  const target = toRaw(this);
  const { has: has2, get: get3 } = getProto(target);
  let hadKey = has2.call(target, key);
  if (!hadKey) {
    key = toRaw(key);
    hadKey = has2.call(target, key);
  } else {
    checkIdentityKeys(target, has2, key);
  }
  const oldValue = get3.call(target, key);
  target.set(key, value);
  if (!hadKey) {
    trigger(target, "add", key, value);
  } else if (hasChanged(value, oldValue)) {
    trigger(target, "set", key, value, oldValue);
  }
  return this;
}
function deleteEntry(key) {
  const target = toRaw(this);
  const { has: has2, get: get3 } = getProto(target);
  let hadKey = has2.call(target, key);
  if (!hadKey) {
    key = toRaw(key);
    hadKey = has2.call(target, key);
  } else {
    checkIdentityKeys(target, has2, key);
  }
  const oldValue = get3 ? get3.call(target, key) : void 0;
  const result = target.delete(key);
  if (hadKey) {
    trigger(target, "delete", key, void 0, oldValue);
  }
  return result;
}
function clear() {
  const target = toRaw(this);
  const hadItems = target.size !== 0;
  const oldTarget = isMap(target) ? new Map(target) : new Set(target) ;
  const result = target.clear();
  if (hadItems) {
    trigger(target, "clear", void 0, void 0, oldTarget);
  }
  return result;
}
function createForEach(isReadonly, isShallow) {
  return function forEach(callback, thisArg) {
    const observed = this;
    const target = observed[
      "__v_raw"
      /* RAW */
    ];
    const rawTarget = toRaw(target);
    const wrap = isShallow ? toShallow : isReadonly ? toReadonly : toReactive;
    !isReadonly && track(rawTarget, "iterate", ITERATE_KEY);
    return target.forEach((value, key) => {
      return callback.call(thisArg, wrap(value), wrap(key), observed);
    });
  };
}
function createIterableMethod(method, isReadonly, isShallow) {
  return function(...args) {
    const target = this[
      "__v_raw"
      /* RAW */
    ];
    const rawTarget = toRaw(target);
    const targetIsMap = isMap(rawTarget);
    const isPair = method === "entries" || method === Symbol.iterator && targetIsMap;
    const isKeyOnly = method === "keys" && targetIsMap;
    const innerIterator = target[method](...args);
    const wrap = isShallow ? toShallow : isReadonly ? toReadonly : toReactive;
    !isReadonly && track(rawTarget, "iterate", isKeyOnly ? MAP_KEY_ITERATE_KEY : ITERATE_KEY);
    return {
      // iterator protocol
      next() {
        const { value, done } = innerIterator.next();
        return done ? { value, done } : {
          value: isPair ? [wrap(value[0]), wrap(value[1])] : wrap(value),
          done
        };
      },
      // iterable protocol
      [Symbol.iterator]() {
        return this;
      }
    };
  };
}
function createReadonlyMethod(type) {
  return function(...args) {
    {
      const key = args[0] ? `on key "${args[0]}" ` : ``;
      console.warn(`${capitalize(type)} operation ${key}failed: target is readonly.`, toRaw(this));
    }
    return type === "delete" ? false : this;
  };
}
function createInstrumentations() {
  const mutableInstrumentations2 = {
    get(key) {
      return get$1(this, key);
    },
    get size() {
      return size(this);
    },
    has: has$1,
    add,
    set: set$1,
    delete: deleteEntry,
    clear,
    forEach: createForEach(false, false)
  };
  const shallowInstrumentations2 = {
    get(key) {
      return get$1(this, key, false, true);
    },
    get size() {
      return size(this);
    },
    has: has$1,
    add,
    set: set$1,
    delete: deleteEntry,
    clear,
    forEach: createForEach(false, true)
  };
  const readonlyInstrumentations2 = {
    get(key) {
      return get$1(this, key, true);
    },
    get size() {
      return size(this, true);
    },
    has(key) {
      return has$1.call(this, key, true);
    },
    add: createReadonlyMethod(
      "add"
      /* ADD */
    ),
    set: createReadonlyMethod(
      "set"
      /* SET */
    ),
    delete: createReadonlyMethod(
      "delete"
      /* DELETE */
    ),
    clear: createReadonlyMethod(
      "clear"
      /* CLEAR */
    ),
    forEach: createForEach(true, false)
  };
  const shallowReadonlyInstrumentations2 = {
    get(key) {
      return get$1(this, key, true, true);
    },
    get size() {
      return size(this, true);
    },
    has(key) {
      return has$1.call(this, key, true);
    },
    add: createReadonlyMethod(
      "add"
      /* ADD */
    ),
    set: createReadonlyMethod(
      "set"
      /* SET */
    ),
    delete: createReadonlyMethod(
      "delete"
      /* DELETE */
    ),
    clear: createReadonlyMethod(
      "clear"
      /* CLEAR */
    ),
    forEach: createForEach(true, true)
  };
  const iteratorMethods = ["keys", "values", "entries", Symbol.iterator];
  iteratorMethods.forEach((method) => {
    mutableInstrumentations2[method] = createIterableMethod(method, false, false);
    readonlyInstrumentations2[method] = createIterableMethod(method, true, false);
    shallowInstrumentations2[method] = createIterableMethod(method, false, true);
    shallowReadonlyInstrumentations2[method] = createIterableMethod(method, true, true);
  });
  return [
    mutableInstrumentations2,
    readonlyInstrumentations2,
    shallowInstrumentations2,
    shallowReadonlyInstrumentations2
  ];
}
var [mutableInstrumentations, readonlyInstrumentations] = /* @__PURE__ */ createInstrumentations();
function createInstrumentationGetter(isReadonly, shallow) {
  const instrumentations = isReadonly ? readonlyInstrumentations : mutableInstrumentations;
  return (target, key, receiver) => {
    if (key === "__v_isReactive") {
      return !isReadonly;
    } else if (key === "__v_isReadonly") {
      return isReadonly;
    } else if (key === "__v_raw") {
      return target;
    }
    return Reflect.get(hasOwn(instrumentations, key) && key in target ? instrumentations : target, key, receiver);
  };
}
var mutableCollectionHandlers = {
  get: /* @__PURE__ */ createInstrumentationGetter(false)
};
var readonlyCollectionHandlers = {
  get: /* @__PURE__ */ createInstrumentationGetter(true)
};
function checkIdentityKeys(target, has2, key) {
  const rawKey = toRaw(key);
  if (rawKey !== key && has2.call(target, rawKey)) {
    const type = toRawType(target);
    console.warn(`Reactive ${type} contains both the raw and reactive versions of the same object${type === `Map` ? ` as keys` : ``}, which can lead to inconsistencies. Avoid differentiating between the raw and reactive versions of an object and only use the reactive version if possible.`);
  }
}
var reactiveMap = /* @__PURE__ */ new WeakMap();
var shallowReactiveMap = /* @__PURE__ */ new WeakMap();
var readonlyMap = /* @__PURE__ */ new WeakMap();
var shallowReadonlyMap = /* @__PURE__ */ new WeakMap();
function targetTypeMap(rawType) {
  switch (rawType) {
    case "Object":
    case "Array":
      return 1;
    case "Map":
    case "Set":
    case "WeakMap":
    case "WeakSet":
      return 2;
    default:
      return 0;
  }
}
function getTargetType(value) {
  return value[
    "__v_skip"
    /* SKIP */
  ] || !Object.isExtensible(value) ? 0 : targetTypeMap(toRawType(value));
}
function reactive2(target) {
  if (target && target[
    "__v_isReadonly"
    /* IS_READONLY */
  ]) {
    return target;
  }
  return createReactiveObject(target, false, mutableHandlers, mutableCollectionHandlers, reactiveMap);
}
function readonly(target) {
  return createReactiveObject(target, true, readonlyHandlers, readonlyCollectionHandlers, readonlyMap);
}
function createReactiveObject(target, isReadonly, baseHandlers, collectionHandlers, proxyMap) {
  if (!isObject(target)) {
    {
      console.warn(`value cannot be made reactive: ${String(target)}`);
    }
    return target;
  }
  if (target[
    "__v_raw"
    /* RAW */
  ] && !(isReadonly && target[
    "__v_isReactive"
    /* IS_REACTIVE */
  ])) {
    return target;
  }
  const existingProxy = proxyMap.get(target);
  if (existingProxy) {
    return existingProxy;
  }
  const targetType = getTargetType(target);
  if (targetType === 0) {
    return target;
  }
  const proxy = new Proxy(target, targetType === 2 ? collectionHandlers : baseHandlers);
  proxyMap.set(target, proxy);
  return proxy;
}
function toRaw(observed) {
  return observed && toRaw(observed[
    "__v_raw"
    /* RAW */
  ]) || observed;
}
function isRef(r) {
  return Boolean(r && r.__v_isRef === true);
}

// packages/alpinejs/src/magics/$nextTick.js
magic("nextTick", () => nextTick);

// packages/alpinejs/src/magics/$dispatch.js
magic("dispatch", (el) => dispatch.bind(dispatch, el));

// packages/alpinejs/src/magics/$watch.js
magic("watch", (el, { evaluateLater: evaluateLater2, cleanup: cleanup2 }) => (key, callback) => {
  let evaluate2 = evaluateLater2(key);
  let getter = () => {
    let value;
    evaluate2((i) => value = i);
    return value;
  };
  let unwatch = watch(getter, callback);
  cleanup2(unwatch);
});

// packages/alpinejs/src/magics/$store.js
magic("store", getStores);

// packages/alpinejs/src/magics/$data.js
magic("data", (el) => scope(el));

// packages/alpinejs/src/magics/$root.js
magic("root", (el) => closestRoot(el));

// packages/alpinejs/src/magics/$refs.js
magic("refs", (el) => {
  if (el._x_refs_proxy)
    return el._x_refs_proxy;
  el._x_refs_proxy = mergeProxies(getArrayOfRefObject(el));
  return el._x_refs_proxy;
});
function getArrayOfRefObject(el) {
  let refObjects = [];
  findClosest(el, (i) => {
    if (i._x_refs)
      refObjects.push(i._x_refs);
  });
  return refObjects;
}

// packages/alpinejs/src/ids.js
var globalIdMemo = {};
function findAndIncrementId(name) {
  if (!globalIdMemo[name])
    globalIdMemo[name] = 0;
  return ++globalIdMemo[name];
}
function closestIdRoot(el, name) {
  return findClosest(el, (element) => {
    if (element._x_ids && element._x_ids[name])
      return true;
  });
}
function setIdRoot(el, name) {
  if (!el._x_ids)
    el._x_ids = {};
  if (!el._x_ids[name])
    el._x_ids[name] = findAndIncrementId(name);
}

// packages/alpinejs/src/magics/$id.js
magic("id", (el, { cleanup: cleanup2 }) => (name, key = null) => {
  let cacheKey = `${name}${key ? `-${key}` : ""}`;
  return cacheIdByNameOnElement(el, cacheKey, cleanup2, () => {
    let root = closestIdRoot(el, name);
    let id = root ? root._x_ids[name] : findAndIncrementId(name);
    return key ? `${name}-${id}-${key}` : `${name}-${id}`;
  });
});
interceptClone((from, to) => {
  if (from._x_id) {
    to._x_id = from._x_id;
  }
});
function cacheIdByNameOnElement(el, cacheKey, cleanup2, callback) {
  if (!el._x_id)
    el._x_id = {};
  if (el._x_id[cacheKey])
    return el._x_id[cacheKey];
  let output = callback();
  el._x_id[cacheKey] = output;
  cleanup2(() => {
    delete el._x_id[cacheKey];
  });
  return output;
}

// packages/alpinejs/src/magics/$el.js
magic("el", (el) => el);

// packages/alpinejs/src/magics/index.js
warnMissingPluginMagic("Focus", "focus", "focus");
warnMissingPluginMagic("Persist", "persist", "persist");
function warnMissingPluginMagic(name, magicName, slug) {
  magic(magicName, (el) => warn(`You can't use [$${magicName}] without first installing the "${name}" plugin here: https://alpinejs.dev/plugins/${slug}`, el));
}

// packages/alpinejs/src/directives/x-modelable.js
directive("modelable", (el, { expression }, { effect: effect3, evaluateLater: evaluateLater2, cleanup: cleanup2 }) => {
  let func = evaluateLater2(expression);
  let innerGet = () => {
    let result;
    func((i) => result = i);
    return result;
  };
  let evaluateInnerSet = evaluateLater2(`${expression} = __placeholder`);
  let innerSet = (val) => evaluateInnerSet(() => {
  }, { scope: { "__placeholder": val } });
  let initialValue = innerGet();
  innerSet(initialValue);
  queueMicrotask(() => {
    if (!el._x_model)
      return;
    el._x_removeModelListeners["default"]();
    let outerGet = el._x_model.get;
    let outerSet = el._x_model.set;
    let releaseEntanglement = entangle(
      {
        get() {
          return outerGet();
        },
        set(value) {
          outerSet(value);
        }
      },
      {
        get() {
          return innerGet();
        },
        set(value) {
          innerSet(value);
        }
      }
    );
    cleanup2(releaseEntanglement);
  });
});

// packages/alpinejs/src/directives/x-teleport.js
directive("teleport", (el, { modifiers, expression }, { cleanup: cleanup2 }) => {
  if (el.tagName.toLowerCase() !== "template")
    warn("x-teleport can only be used on a <template> tag", el);
  let target = getTarget(expression);
  let clone2 = el.content.cloneNode(true).firstElementChild;
  el._x_teleport = clone2;
  clone2._x_teleportBack = el;
  el.setAttribute("data-teleport-template", true);
  clone2.setAttribute("data-teleport-target", true);
  if (el._x_forwardEvents) {
    el._x_forwardEvents.forEach((eventName) => {
      clone2.addEventListener(eventName, (e) => {
        e.stopPropagation();
        el.dispatchEvent(new e.constructor(e.type, e));
      });
    });
  }
  addScopeToNode(clone2, {}, el);
  let placeInDom = (clone3, target2, modifiers2) => {
    if (modifiers2.includes("prepend")) {
      target2.parentNode.insertBefore(clone3, target2);
    } else if (modifiers2.includes("append")) {
      target2.parentNode.insertBefore(clone3, target2.nextSibling);
    } else {
      target2.appendChild(clone3);
    }
  };
  mutateDom(() => {
    placeInDom(clone2, target, modifiers);
    skipDuringClone(() => {
      initTree(clone2);
    })();
  });
  el._x_teleportPutBack = () => {
    let target2 = getTarget(expression);
    mutateDom(() => {
      placeInDom(el._x_teleport, target2, modifiers);
    });
  };
  cleanup2(
    () => mutateDom(() => {
      clone2.remove();
      destroyTree(clone2);
    })
  );
});
var teleportContainerDuringClone = document.createElement("div");
function getTarget(expression) {
  let target = skipDuringClone(() => {
    return document.querySelector(expression);
  }, () => {
    return teleportContainerDuringClone;
  })();
  if (!target)
    warn(`Cannot find x-teleport element for selector: "${expression}"`);
  return target;
}

// packages/alpinejs/src/directives/x-ignore.js
var handler = () => {
};
handler.inline = (el, { modifiers }, { cleanup: cleanup2 }) => {
  modifiers.includes("self") ? el._x_ignoreSelf = true : el._x_ignore = true;
  cleanup2(() => {
    modifiers.includes("self") ? delete el._x_ignoreSelf : delete el._x_ignore;
  });
};
directive("ignore", handler);

// packages/alpinejs/src/directives/x-effect.js
directive("effect", skipDuringClone((el, { expression }, { effect: effect3 }) => {
  effect3(evaluateLater(el, expression));
}));

// packages/alpinejs/src/utils/on.js
function on(el, event, modifiers, callback) {
  let listenerTarget = el;
  let handler4 = (e) => callback(e);
  let options = {};
  let wrapHandler = (callback2, wrapper) => (e) => wrapper(callback2, e);
  if (modifiers.includes("dot"))
    event = dotSyntax(event);
  if (modifiers.includes("camel"))
    event = camelCase2(event);
  if (modifiers.includes("passive"))
    options.passive = true;
  if (modifiers.includes("capture"))
    options.capture = true;
  if (modifiers.includes("window"))
    listenerTarget = window;
  if (modifiers.includes("document"))
    listenerTarget = document;
  if (modifiers.includes("debounce")) {
    let nextModifier = modifiers[modifiers.indexOf("debounce") + 1] || "invalid-wait";
    let wait = isNumeric(nextModifier.split("ms")[0]) ? Number(nextModifier.split("ms")[0]) : 250;
    handler4 = debounce(handler4, wait);
  }
  if (modifiers.includes("throttle")) {
    let nextModifier = modifiers[modifiers.indexOf("throttle") + 1] || "invalid-wait";
    let wait = isNumeric(nextModifier.split("ms")[0]) ? Number(nextModifier.split("ms")[0]) : 250;
    handler4 = throttle(handler4, wait);
  }
  if (modifiers.includes("prevent"))
    handler4 = wrapHandler(handler4, (next, e) => {
      e.preventDefault();
      next(e);
    });
  if (modifiers.includes("stop"))
    handler4 = wrapHandler(handler4, (next, e) => {
      e.stopPropagation();
      next(e);
    });
  if (modifiers.includes("once")) {
    handler4 = wrapHandler(handler4, (next, e) => {
      next(e);
      listenerTarget.removeEventListener(event, handler4, options);
    });
  }
  if (modifiers.includes("away") || modifiers.includes("outside")) {
    listenerTarget = document;
    handler4 = wrapHandler(handler4, (next, e) => {
      if (el.contains(e.target))
        return;
      if (e.target.isConnected === false)
        return;
      if (el.offsetWidth < 1 && el.offsetHeight < 1)
        return;
      if (el._x_isShown === false)
        return;
      next(e);
    });
  }
  if (modifiers.includes("self"))
    handler4 = wrapHandler(handler4, (next, e) => {
      e.target === el && next(e);
    });
  if (isKeyEvent(event) || isClickEvent(event)) {
    handler4 = wrapHandler(handler4, (next, e) => {
      if (isListeningForASpecificKeyThatHasntBeenPressed(e, modifiers)) {
        return;
      }
      next(e);
    });
  }
  listenerTarget.addEventListener(event, handler4, options);
  return () => {
    listenerTarget.removeEventListener(event, handler4, options);
  };
}
function dotSyntax(subject) {
  return subject.replace(/-/g, ".");
}
function camelCase2(subject) {
  return subject.toLowerCase().replace(/-(\w)/g, (match, char) => char.toUpperCase());
}
function isNumeric(subject) {
  return !Array.isArray(subject) && !isNaN(subject);
}
function kebabCase2(subject) {
  if ([" ", "_"].includes(
    subject
  ))
    return subject;
  return subject.replace(/([a-z])([A-Z])/g, "$1-$2").replace(/[_\s]/, "-").toLowerCase();
}
function isKeyEvent(event) {
  return ["keydown", "keyup"].includes(event);
}
function isClickEvent(event) {
  return ["contextmenu", "click", "mouse"].some((i) => event.includes(i));
}
function isListeningForASpecificKeyThatHasntBeenPressed(e, modifiers) {
  let keyModifiers = modifiers.filter((i) => {
    return !["window", "document", "prevent", "stop", "once", "capture", "self", "away", "outside", "passive", "preserve-scroll"].includes(i);
  });
  if (keyModifiers.includes("debounce")) {
    let debounceIndex = keyModifiers.indexOf("debounce");
    keyModifiers.splice(debounceIndex, isNumeric((keyModifiers[debounceIndex + 1] || "invalid-wait").split("ms")[0]) ? 2 : 1);
  }
  if (keyModifiers.includes("throttle")) {
    let debounceIndex = keyModifiers.indexOf("throttle");
    keyModifiers.splice(debounceIndex, isNumeric((keyModifiers[debounceIndex + 1] || "invalid-wait").split("ms")[0]) ? 2 : 1);
  }
  if (keyModifiers.length === 0)
    return false;
  if (keyModifiers.length === 1 && keyToModifiers(e.key).includes(keyModifiers[0]))
    return false;
  const systemKeyModifiers = ["ctrl", "shift", "alt", "meta", "cmd", "super"];
  const selectedSystemKeyModifiers = systemKeyModifiers.filter((modifier) => keyModifiers.includes(modifier));
  keyModifiers = keyModifiers.filter((i) => !selectedSystemKeyModifiers.includes(i));
  if (selectedSystemKeyModifiers.length > 0) {
    const activelyPressedKeyModifiers = selectedSystemKeyModifiers.filter((modifier) => {
      if (modifier === "cmd" || modifier === "super")
        modifier = "meta";
      return e[`${modifier}Key`];
    });
    if (activelyPressedKeyModifiers.length === selectedSystemKeyModifiers.length) {
      if (isClickEvent(e.type))
        return false;
      if (keyToModifiers(e.key).includes(keyModifiers[0]))
        return false;
    }
  }
  return true;
}
function keyToModifiers(key) {
  if (!key)
    return [];
  key = kebabCase2(key);
  let modifierToKeyMap = {
    "ctrl": "control",
    "slash": "/",
    "space": " ",
    "spacebar": " ",
    "cmd": "meta",
    "esc": "escape",
    "up": "arrow-up",
    "down": "arrow-down",
    "left": "arrow-left",
    "right": "arrow-right",
    "period": ".",
    "comma": ",",
    "equal": "=",
    "minus": "-",
    "underscore": "_"
  };
  modifierToKeyMap[key] = key;
  return Object.keys(modifierToKeyMap).map((modifier) => {
    if (modifierToKeyMap[modifier] === key)
      return modifier;
  }).filter((modifier) => modifier);
}

// packages/alpinejs/src/directives/x-model.js
directive("model", (el, { modifiers, expression }, { effect: effect3, cleanup: cleanup2 }) => {
  let scopeTarget = el;
  if (modifiers.includes("parent")) {
    scopeTarget = el.parentNode;
  }
  let evaluateGet = evaluateLater(scopeTarget, expression);
  let evaluateSet;
  if (typeof expression === "string") {
    evaluateSet = evaluateLater(scopeTarget, `${expression} = __placeholder`);
  } else if (typeof expression === "function" && typeof expression() === "string") {
    evaluateSet = evaluateLater(scopeTarget, `${expression()} = __placeholder`);
  } else {
    evaluateSet = () => {
    };
  }
  let getValue = () => {
    let result;
    evaluateGet((value) => result = value);
    return isGetterSetter(result) ? result.get() : result;
  };
  let setValue = (value) => {
    let result;
    evaluateGet((value2) => result = value2);
    if (isGetterSetter(result)) {
      result.set(value);
    } else {
      evaluateSet(() => {
      }, {
        scope: { "__placeholder": value }
      });
    }
  };
  if (typeof expression === "string" && el.type === "radio") {
    mutateDom(() => {
      if (!el.hasAttribute("name"))
        el.setAttribute("name", expression);
    });
  }
  let event = el.tagName.toLowerCase() === "select" || ["checkbox", "radio"].includes(el.type) || modifiers.includes("lazy") ? "change" : "input";
  let removeListener = isCloning ? () => {
  } : on(el, event, modifiers, (e) => {
    setValue(getInputValue(el, modifiers, e, getValue()));
  });
  if (modifiers.includes("fill")) {
    if ([void 0, null, ""].includes(getValue()) || isCheckbox(el) && Array.isArray(getValue()) || el.tagName.toLowerCase() === "select" && el.multiple) {
      setValue(
        getInputValue(el, modifiers, { target: el }, getValue())
      );
    }
  }
  if (!el._x_removeModelListeners)
    el._x_removeModelListeners = {};
  el._x_removeModelListeners["default"] = removeListener;
  cleanup2(() => el._x_removeModelListeners["default"]());
  if (el.form) {
    let removeResetListener = on(el.form, "reset", [], (e) => {
      nextTick(() => el._x_model && el._x_model.set(getInputValue(el, modifiers, { target: el }, getValue())));
    });
    cleanup2(() => removeResetListener());
  }
  el._x_model = {
    get() {
      return getValue();
    },
    set(value) {
      setValue(value);
    }
  };
  el._x_forceModelUpdate = (value) => {
    if (value === void 0 && typeof expression === "string" && expression.match(/\./))
      value = "";
    window.fromModel = true;
    mutateDom(() => bind(el, "value", value));
    delete window.fromModel;
  };
  effect3(() => {
    let value = getValue();
    if (modifiers.includes("unintrusive") && document.activeElement.isSameNode(el))
      return;
    el._x_forceModelUpdate(value);
  });
});
function getInputValue(el, modifiers, event, currentValue) {
  return mutateDom(() => {
    if (event instanceof CustomEvent && event.detail !== void 0)
      return event.detail !== null && event.detail !== void 0 ? event.detail : event.target.value;
    else if (isCheckbox(el)) {
      if (Array.isArray(currentValue)) {
        let newValue = null;
        if (modifiers.includes("number")) {
          newValue = safeParseNumber(event.target.value);
        } else if (modifiers.includes("boolean")) {
          newValue = safeParseBoolean(event.target.value);
        } else {
          newValue = event.target.value;
        }
        return event.target.checked ? currentValue.includes(newValue) ? currentValue : currentValue.concat([newValue]) : currentValue.filter((el2) => !checkedAttrLooseCompare2(el2, newValue));
      } else {
        return event.target.checked;
      }
    } else if (el.tagName.toLowerCase() === "select" && el.multiple) {
      if (modifiers.includes("number")) {
        return Array.from(event.target.selectedOptions).map((option) => {
          let rawValue = option.value || option.text;
          return safeParseNumber(rawValue);
        });
      } else if (modifiers.includes("boolean")) {
        return Array.from(event.target.selectedOptions).map((option) => {
          let rawValue = option.value || option.text;
          return safeParseBoolean(rawValue);
        });
      }
      return Array.from(event.target.selectedOptions).map((option) => {
        return option.value || option.text;
      });
    } else {
      let newValue;
      if (isRadio(el)) {
        if (event.target.checked) {
          newValue = event.target.value;
        } else {
          newValue = currentValue;
        }
      } else {
        newValue = event.target.value;
      }
      if (modifiers.includes("number")) {
        return safeParseNumber(newValue);
      } else if (modifiers.includes("boolean")) {
        return safeParseBoolean(newValue);
      } else if (modifiers.includes("trim")) {
        return newValue.trim();
      } else {
        return newValue;
      }
    }
  });
}
function safeParseNumber(rawValue) {
  let number = rawValue ? parseFloat(rawValue) : null;
  return isNumeric2(number) ? number : rawValue;
}
function checkedAttrLooseCompare2(valueA, valueB) {
  return valueA == valueB;
}
function isNumeric2(subject) {
  return !Array.isArray(subject) && !isNaN(subject);
}
function isGetterSetter(value) {
  return value !== null && typeof value === "object" && typeof value.get === "function" && typeof value.set === "function";
}

// packages/alpinejs/src/directives/x-cloak.js
directive("cloak", (el) => queueMicrotask(() => mutateDom(() => el.removeAttribute(prefix("cloak")))));

// packages/alpinejs/src/directives/x-init.js
addInitSelector(() => `[${prefix("init")}]`);
directive("init", skipDuringClone((el, { expression }, { evaluate: evaluate2 }) => {
  if (typeof expression === "string") {
    return !!expression.trim() && evaluate2(expression, {}, false);
  }
  return evaluate2(expression, {}, false);
}));

// packages/alpinejs/src/directives/x-text.js
directive("text", (el, { expression }, { effect: effect3, evaluateLater: evaluateLater2 }) => {
  let evaluate2 = evaluateLater2(expression);
  effect3(() => {
    evaluate2((value) => {
      mutateDom(() => {
        el.textContent = value;
      });
    });
  });
});

// packages/alpinejs/src/directives/x-html.js
directive("html", (el, { expression }, { effect: effect3, evaluateLater: evaluateLater2 }) => {
  let evaluate2 = evaluateLater2(expression);
  effect3(() => {
    evaluate2((value) => {
      mutateDom(() => {
        el.innerHTML = value;
        el._x_ignoreSelf = true;
        initTree(el);
        delete el._x_ignoreSelf;
      });
    });
  });
});

// packages/alpinejs/src/directives/x-bind.js
mapAttributes(startingWith(":", into(prefix("bind:"))));
var handler2 = (el, { value, modifiers, expression, original }, { effect: effect3, cleanup: cleanup2 }) => {
  if (!value) {
    let bindingProviders = {};
    injectBindingProviders(bindingProviders);
    let getBindings = evaluateLater(el, expression);
    getBindings((bindings) => {
      applyBindingsObject(el, bindings, original);
    }, { scope: bindingProviders });
    return;
  }
  if (value === "key")
    return storeKeyForXFor(el, expression);
  if (el._x_inlineBindings && el._x_inlineBindings[value] && el._x_inlineBindings[value].extract) {
    return;
  }
  let evaluate2 = evaluateLater(el, expression);
  effect3(() => evaluate2((result) => {
    if (result === void 0 && typeof expression === "string" && expression.match(/\./)) {
      result = "";
    }
    mutateDom(() => bind(el, value, result, modifiers));
  }));
  cleanup2(() => {
    el._x_undoAddedClasses && el._x_undoAddedClasses();
    el._x_undoAddedStyles && el._x_undoAddedStyles();
  });
};
handler2.inline = (el, { value, modifiers, expression }) => {
  if (!value)
    return;
  if (!el._x_inlineBindings)
    el._x_inlineBindings = {};
  el._x_inlineBindings[value] = { expression, extract: false };
};
directive("bind", handler2);
function storeKeyForXFor(el, expression) {
  el._x_keyExpression = expression;
}

// packages/alpinejs/src/directives/x-data.js
addRootSelector(() => `[${prefix("data")}]`);
directive("data", (el, { expression }, { cleanup: cleanup2 }) => {
  if (shouldSkipRegisteringDataDuringClone(el))
    return;
  expression = expression === "" ? "{}" : expression;
  let magicContext = {};
  injectMagics(magicContext, el);
  let dataProviderContext = {};
  injectDataProviders(dataProviderContext, magicContext);
  let data2 = evaluate(el, expression, { scope: dataProviderContext });
  if (data2 === void 0 || data2 === true)
    data2 = {};
  injectMagics(data2, el);
  let reactiveData = reactive(data2);
  initInterceptors(reactiveData);
  let undo = addScopeToNode(el, reactiveData);
  reactiveData["init"] && evaluate(el, reactiveData["init"]);
  cleanup2(() => {
    reactiveData["destroy"] && evaluate(el, reactiveData["destroy"]);
    undo();
  });
});
interceptClone((from, to) => {
  if (from._x_dataStack) {
    to._x_dataStack = from._x_dataStack;
    to.setAttribute("data-has-alpine-state", true);
  }
});
function shouldSkipRegisteringDataDuringClone(el) {
  if (!isCloning)
    return false;
  if (isCloningLegacy)
    return true;
  return el.hasAttribute("data-has-alpine-state");
}

// packages/alpinejs/src/directives/x-show.js
directive("show", (el, { modifiers, expression }, { effect: effect3 }) => {
  let evaluate2 = evaluateLater(el, expression);
  if (!el._x_doHide)
    el._x_doHide = () => {
      mutateDom(() => {
        el.style.setProperty("display", "none", modifiers.includes("important") ? "important" : void 0);
      });
    };
  if (!el._x_doShow)
    el._x_doShow = () => {
      mutateDom(() => {
        if (el.style.length === 1 && el.style.display === "none") {
          el.removeAttribute("style");
        } else {
          el.style.removeProperty("display");
        }
      });
    };
  let hide = () => {
    el._x_doHide();
    el._x_isShown = false;
  };
  let show = () => {
    el._x_doShow();
    el._x_isShown = true;
  };
  let clickAwayCompatibleShow = () => setTimeout(show);
  let toggle = once(
    (value) => value ? show() : hide(),
    (value) => {
      if (typeof el._x_toggleAndCascadeWithTransitions === "function") {
        el._x_toggleAndCascadeWithTransitions(el, value, show, hide);
      } else {
        value ? clickAwayCompatibleShow() : hide();
      }
    }
  );
  let oldValue;
  let firstTime = true;
  effect3(() => evaluate2((value) => {
    if (!firstTime && value === oldValue)
      return;
    if (modifiers.includes("immediate"))
      value ? clickAwayCompatibleShow() : hide();
    toggle(value);
    oldValue = value;
    firstTime = false;
  }));
});

// packages/alpinejs/src/directives/x-for.js
directive("for", (el, { expression }, { effect: effect3, cleanup: cleanup2 }) => {
  let iteratorNames = parseForExpression(expression);
  let evaluateItems = evaluateLater(el, iteratorNames.items);
  let evaluateKey = evaluateLater(
    el,
    // the x-bind:key expression is stored for our use instead of evaluated.
    el._x_keyExpression || "index"
  );
  el._x_prevKeys = [];
  el._x_lookup = {};
  effect3(() => loop(el, iteratorNames, evaluateItems, evaluateKey));
  cleanup2(() => {
    Object.values(el._x_lookup).forEach((el2) => mutateDom(
      () => {
        destroyTree(el2);
        el2.remove();
      }
    ));
    delete el._x_prevKeys;
    delete el._x_lookup;
  });
});
function loop(el, iteratorNames, evaluateItems, evaluateKey) {
  let isObject2 = (i) => typeof i === "object" && !Array.isArray(i);
  let templateEl = el;
  evaluateItems((items) => {
    if (isNumeric3(items) && items >= 0) {
      items = Array.from(Array(items).keys(), (i) => i + 1);
    }
    if (items === void 0)
      items = [];
    let lookup = el._x_lookup;
    let prevKeys = el._x_prevKeys;
    let scopes = [];
    let keys = [];
    if (isObject2(items)) {
      items = Object.entries(items).map(([key, value]) => {
        let scope2 = getIterationScopeVariables(iteratorNames, value, key, items);
        evaluateKey((value2) => {
          if (keys.includes(value2))
            warn("Duplicate key on x-for", el);
          keys.push(value2);
        }, { scope: { index: key, ...scope2 } });
        scopes.push(scope2);
      });
    } else {
      for (let i = 0; i < items.length; i++) {
        let scope2 = getIterationScopeVariables(iteratorNames, items[i], i, items);
        evaluateKey((value) => {
          if (keys.includes(value))
            warn("Duplicate key on x-for", el);
          keys.push(value);
        }, { scope: { index: i, ...scope2 } });
        scopes.push(scope2);
      }
    }
    let adds = [];
    let moves = [];
    let removes = [];
    let sames = [];
    for (let i = 0; i < prevKeys.length; i++) {
      let key = prevKeys[i];
      if (keys.indexOf(key) === -1)
        removes.push(key);
    }
    prevKeys = prevKeys.filter((key) => !removes.includes(key));
    let lastKey = "template";
    for (let i = 0; i < keys.length; i++) {
      let key = keys[i];
      let prevIndex = prevKeys.indexOf(key);
      if (prevIndex === -1) {
        prevKeys.splice(i, 0, key);
        adds.push([lastKey, i]);
      } else if (prevIndex !== i) {
        let keyInSpot = prevKeys.splice(i, 1)[0];
        let keyForSpot = prevKeys.splice(prevIndex - 1, 1)[0];
        prevKeys.splice(i, 0, keyForSpot);
        prevKeys.splice(prevIndex, 0, keyInSpot);
        moves.push([keyInSpot, keyForSpot]);
      } else {
        sames.push(key);
      }
      lastKey = key;
    }
    for (let i = 0; i < removes.length; i++) {
      let key = removes[i];
      if (!(key in lookup))
        continue;
      mutateDom(() => {
        destroyTree(lookup[key]);
        lookup[key].remove();
      });
      delete lookup[key];
    }
    for (let i = 0; i < moves.length; i++) {
      let [keyInSpot, keyForSpot] = moves[i];
      let elInSpot = lookup[keyInSpot];
      let elForSpot = lookup[keyForSpot];
      let marker = document.createElement("div");
      mutateDom(() => {
        if (!elForSpot)
          warn(`x-for ":key" is undefined or invalid`, templateEl, keyForSpot, lookup);
        elForSpot.after(marker);
        elInSpot.after(elForSpot);
        elForSpot._x_currentIfEl && elForSpot.after(elForSpot._x_currentIfEl);
        marker.before(elInSpot);
        elInSpot._x_currentIfEl && elInSpot.after(elInSpot._x_currentIfEl);
        marker.remove();
      });
      elForSpot._x_refreshXForScope(scopes[keys.indexOf(keyForSpot)]);
    }
    for (let i = 0; i < adds.length; i++) {
      let [lastKey2, index] = adds[i];
      let lastEl = lastKey2 === "template" ? templateEl : lookup[lastKey2];
      if (lastEl._x_currentIfEl)
        lastEl = lastEl._x_currentIfEl;
      let scope2 = scopes[index];
      let key = keys[index];
      let clone2 = document.importNode(templateEl.content, true).firstElementChild;
      let reactiveScope = reactive(scope2);
      addScopeToNode(clone2, reactiveScope, templateEl);
      clone2._x_refreshXForScope = (newScope) => {
        Object.entries(newScope).forEach(([key2, value]) => {
          reactiveScope[key2] = value;
        });
      };
      mutateDom(() => {
        lastEl.after(clone2);
        skipDuringClone(() => initTree(clone2))();
      });
      if (typeof key === "object") {
        warn("x-for key cannot be an object, it must be a string or an integer", templateEl);
      }
      lookup[key] = clone2;
    }
    for (let i = 0; i < sames.length; i++) {
      lookup[sames[i]]._x_refreshXForScope(scopes[keys.indexOf(sames[i])]);
    }
    templateEl._x_prevKeys = keys;
  });
}
function parseForExpression(expression) {
  let forIteratorRE = /,([^,\}\]]*)(?:,([^,\}\]]*))?$/;
  let stripParensRE = /^\s*\(|\)\s*$/g;
  let forAliasRE = /([\s\S]*?)\s+(?:in|of)\s+([\s\S]*)/;
  let inMatch = expression.match(forAliasRE);
  if (!inMatch)
    return;
  let res = {};
  res.items = inMatch[2].trim();
  let item = inMatch[1].replace(stripParensRE, "").trim();
  let iteratorMatch = item.match(forIteratorRE);
  if (iteratorMatch) {
    res.item = item.replace(forIteratorRE, "").trim();
    res.index = iteratorMatch[1].trim();
    if (iteratorMatch[2]) {
      res.collection = iteratorMatch[2].trim();
    }
  } else {
    res.item = item;
  }
  return res;
}
function getIterationScopeVariables(iteratorNames, item, index, items) {
  let scopeVariables = {};
  if (/^\[.*\]$/.test(iteratorNames.item) && Array.isArray(item)) {
    let names = iteratorNames.item.replace("[", "").replace("]", "").split(",").map((i) => i.trim());
    names.forEach((name, i) => {
      scopeVariables[name] = item[i];
    });
  } else if (/^\{.*\}$/.test(iteratorNames.item) && !Array.isArray(item) && typeof item === "object") {
    let names = iteratorNames.item.replace("{", "").replace("}", "").split(",").map((i) => i.trim());
    names.forEach((name) => {
      scopeVariables[name] = item[name];
    });
  } else {
    scopeVariables[iteratorNames.item] = item;
  }
  if (iteratorNames.index)
    scopeVariables[iteratorNames.index] = index;
  if (iteratorNames.collection)
    scopeVariables[iteratorNames.collection] = items;
  return scopeVariables;
}
function isNumeric3(subject) {
  return !Array.isArray(subject) && !isNaN(subject);
}

// packages/alpinejs/src/directives/x-ref.js
function handler3() {
}
handler3.inline = (el, { expression }, { cleanup: cleanup2 }) => {
  let root = closestRoot(el);
  if (!root._x_refs)
    root._x_refs = {};
  root._x_refs[expression] = el;
  cleanup2(() => delete root._x_refs[expression]);
};
directive("ref", handler3);

// packages/alpinejs/src/directives/x-if.js
directive("if", (el, { expression }, { effect: effect3, cleanup: cleanup2 }) => {
  if (el.tagName.toLowerCase() !== "template")
    warn("x-if can only be used on a <template> tag", el);
  let evaluate2 = evaluateLater(el, expression);
  let show = () => {
    if (el._x_currentIfEl)
      return el._x_currentIfEl;
    let clone2 = el.content.cloneNode(true).firstElementChild;
    addScopeToNode(clone2, {}, el);
    mutateDom(() => {
      el.after(clone2);
      skipDuringClone(() => initTree(clone2))();
    });
    el._x_currentIfEl = clone2;
    el._x_undoIf = () => {
      mutateDom(() => {
        destroyTree(clone2);
        clone2.remove();
      });
      delete el._x_currentIfEl;
    };
    return clone2;
  };
  let hide = () => {
    if (!el._x_undoIf)
      return;
    el._x_undoIf();
    delete el._x_undoIf;
  };
  effect3(() => evaluate2((value) => {
    value ? show() : hide();
  }));
  cleanup2(() => el._x_undoIf && el._x_undoIf());
});

// packages/alpinejs/src/directives/x-id.js
directive("id", (el, { expression }, { evaluate: evaluate2 }) => {
  let names = evaluate2(expression);
  names.forEach((name) => setIdRoot(el, name));
});
interceptClone((from, to) => {
  if (from._x_ids) {
    to._x_ids = from._x_ids;
  }
});

// packages/alpinejs/src/directives/x-on.js
mapAttributes(startingWith("@", into(prefix("on:"))));
directive("on", skipDuringClone((el, { value, modifiers, expression }, { cleanup: cleanup2 }) => {
  let evaluate2 = expression ? evaluateLater(el, expression) : () => {
  };
  if (el.tagName.toLowerCase() === "template") {
    if (!el._x_forwardEvents)
      el._x_forwardEvents = [];
    if (!el._x_forwardEvents.includes(value))
      el._x_forwardEvents.push(value);
  }
  let removeListener = on(el, value, modifiers, (e) => {
    evaluate2(() => {
    }, { scope: { "$event": e }, params: [e] });
  });
  cleanup2(() => removeListener());
}));

// packages/alpinejs/src/directives/index.js
warnMissingPluginDirective("Collapse", "collapse", "collapse");
warnMissingPluginDirective("Intersect", "intersect", "intersect");
warnMissingPluginDirective("Focus", "trap", "focus");
warnMissingPluginDirective("Mask", "mask", "mask");
function warnMissingPluginDirective(name, directiveName, slug) {
  directive(directiveName, (el) => warn(`You can't use [x-${directiveName}] without first installing the "${name}" plugin here: https://alpinejs.dev/plugins/${slug}`, el));
}

// packages/alpinejs/src/index.js
alpine_default.setEvaluator(normalEvaluator);
alpine_default.setRawEvaluator(normalRawEvaluator);
alpine_default.setReactivityEngine({ reactive: reactive2, effect: effect2, release: stop, raw: toRaw });
var src_default = alpine_default;

// packages/alpinejs/builds/module.js
var module_default = src_default;

var htmx$1 = (function() {

  // Public API
  const htmx = {
    // Tsc madness here, assigning the functions directly results in an invalid TypeScript output, but reassigning is fine
    /* Event processing */
    /** @type {typeof onLoadHelper} */
    onLoad: null,
    /** @type {typeof processNode} */
    process: null,
    /** @type {typeof addEventListenerImpl} */
    on: null,
    /** @type {typeof removeEventListenerImpl} */
    off: null,
    /** @type {typeof triggerEvent} */
    trigger: null,
    /** @type {typeof ajaxHelper} */
    ajax: null,
    /* DOM querying helpers */
    /** @type {typeof find} */
    find: null,
    /** @type {typeof findAll} */
    findAll: null,
    /** @type {typeof closest} */
    closest: null,
    /**
     * Returns the input values that would resolve for a given element via the htmx value resolution mechanism
     *
     * @see https://htmx.org/api/#values
     *
     * @param {Element} elt the element to resolve values on
     * @param {HttpVerb} type the request type (e.g. **get** or **post**) non-GET's will include the enclosing form of the element. Defaults to **post**
     * @returns {Object}
     */
    values: function(elt, type) {
      const inputValues = getInputValues(elt, type || 'post');
      return inputValues.values
    },
    /* DOM manipulation helpers */
    /** @type {typeof removeElement} */
    remove: null,
    /** @type {typeof addClassToElement} */
    addClass: null,
    /** @type {typeof removeClassFromElement} */
    removeClass: null,
    /** @type {typeof toggleClassOnElement} */
    toggleClass: null,
    /** @type {typeof takeClassForElement} */
    takeClass: null,
    /** @type {typeof swap} */
    swap: null,
    /* Extension entrypoints */
    /** @type {typeof defineExtension} */
    defineExtension: null,
    /** @type {typeof removeExtension} */
    removeExtension: null,
    /* Debugging */
    /** @type {typeof logAll} */
    logAll: null,
    /** @type {typeof logNone} */
    logNone: null,
    /* Debugging */
    /**
     * The logger htmx uses to log with
     *
     * @see https://htmx.org/api/#logger
     */
    logger: null,
    /**
     * A property holding the configuration htmx uses at runtime.
     *
     * Note that using a [meta tag](https://htmx.org/docs/#config) is the preferred mechanism for setting these properties.
     *
     * @see https://htmx.org/api/#config
     */
    config: {
      /**
       * Whether to use history.
       * @type boolean
       * @default true
       */
      historyEnabled: true,
      /**
       * The number of pages to keep in **sessionStorage** for history support.
       * @type number
       * @default 10
       */
      historyCacheSize: 10,
      /**
       * @type boolean
       * @default false
       */
      refreshOnHistoryMiss: false,
      /**
       * The default swap style to use if **[hx-swap](https://htmx.org/attributes/hx-swap)** is omitted.
       * @type HtmxSwapStyle
       * @default 'innerHTML'
       */
      defaultSwapStyle: 'innerHTML',
      /**
       * The default delay between receiving a response from the server and doing the swap.
       * @type number
       * @default 0
       */
      defaultSwapDelay: 0,
      /**
       * The default delay between completing the content swap and settling attributes.
       * @type number
       * @default 20
       */
      defaultSettleDelay: 20,
      /**
       * If true, htmx will inject a small amount of CSS into the page to make indicators invisible unless the **htmx-indicator** class is present.
       * @type boolean
       * @default true
       */
      includeIndicatorStyles: true,
      /**
       * The class to place on indicators when a request is in flight.
       * @type string
       * @default 'htmx-indicator'
       */
      indicatorClass: 'htmx-indicator',
      /**
       * The class to place on triggering elements when a request is in flight.
       * @type string
       * @default 'htmx-request'
       */
      requestClass: 'htmx-request',
      /**
       * The class to temporarily place on elements that htmx has added to the DOM.
       * @type string
       * @default 'htmx-added'
       */
      addedClass: 'htmx-added',
      /**
       * The class to place on target elements when htmx is in the settling phase.
       * @type string
       * @default 'htmx-settling'
       */
      settlingClass: 'htmx-settling',
      /**
       * The class to place on target elements when htmx is in the swapping phase.
       * @type string
       * @default 'htmx-swapping'
       */
      swappingClass: 'htmx-swapping',
      /**
       * Allows the use of eval-like functionality in htmx, to enable **hx-vars**, trigger conditions & script tag evaluation. Can be set to **false** for CSP compatibility.
       * @type boolean
       * @default true
       */
      allowEval: true,
      /**
       * If set to false, disables the interpretation of script tags.
       * @type boolean
       * @default true
       */
      allowScriptTags: true,
      /**
       * If set, the nonce will be added to inline scripts.
       * @type string
       * @default ''
       */
      inlineScriptNonce: '',
      /**
       * If set, the nonce will be added to inline styles.
       * @type string
       * @default ''
       */
      inlineStyleNonce: '',
      /**
       * The attributes to settle during the settling phase.
       * @type string[]
       * @default ['class', 'style', 'width', 'height']
       */
      attributesToSettle: ['class', 'style', 'width', 'height'],
      /**
       * Allow cross-site Access-Control requests using credentials such as cookies, authorization headers or TLS client certificates.
       * @type boolean
       * @default false
       */
      withCredentials: false,
      /**
       * @type number
       * @default 0
       */
      timeout: 0,
      /**
       * The default implementation of **getWebSocketReconnectDelay** for reconnecting after unexpected connection loss by the event code **Abnormal Closure**, **Service Restart** or **Try Again Later**.
       * @type {'full-jitter' | ((retryCount:number) => number)}
       * @default "full-jitter"
       */
      wsReconnectDelay: 'full-jitter',
      /**
       * The type of binary data being received over the WebSocket connection
       * @type BinaryType
       * @default 'blob'
       */
      wsBinaryType: 'blob',
      /**
       * @type string
       * @default '[hx-disable], [data-hx-disable]'
       */
      disableSelector: '[hx-disable], [data-hx-disable]',
      /**
       * @type {'auto' | 'instant' | 'smooth'}
       * @default 'instant'
       */
      scrollBehavior: 'instant',
      /**
       * If the focused element should be scrolled into view.
       * @type boolean
       * @default false
       */
      defaultFocusScroll: false,
      /**
       * If set to true htmx will include a cache-busting parameter in GET requests to avoid caching partial responses by the browser
       * @type boolean
       * @default false
       */
      getCacheBusterParam: false,
      /**
       * If set to true, htmx will use the View Transition API when swapping in new content.
       * @type boolean
       * @default false
       */
      globalViewTransitions: false,
      /**
       * htmx will format requests with these methods by encoding their parameters in the URL, not the request body
       * @type {(HttpVerb)[]}
       * @default ['get', 'delete']
       */
      methodsThatUseUrlParams: ['get', 'delete'],
      /**
       * If set to true, disables htmx-based requests to non-origin hosts.
       * @type boolean
       * @default false
       */
      selfRequestsOnly: true,
      /**
       * If set to true htmx will not update the title of the document when a title tag is found in new content
       * @type boolean
       * @default false
       */
      ignoreTitle: false,
      /**
       * Whether the target of a boosted element is scrolled into the viewport.
       * @type boolean
       * @default true
       */
      scrollIntoViewOnBoost: true,
      /**
       * The cache to store evaluated trigger specifications into.
       * You may define a simple object to use a never-clearing cache, or implement your own system using a [proxy object](https://developer.mozilla.org/docs/Web/JavaScript/Reference/Global_Objects/Proxy)
       * @type {Object|null}
       * @default null
       */
      triggerSpecsCache: null,
      /** @type boolean */
      disableInheritance: false,
      /** @type HtmxResponseHandlingConfig[] */
      responseHandling: [
        { code: '204', swap: false },
        { code: '[23]..', swap: true },
        { code: '[45]..', swap: false, error: true }
      ],
      /**
       * Whether to process OOB swaps on elements that are nested within the main response element.
       * @type boolean
       * @default true
       */
      allowNestedOobSwaps: true,
      /**
       * Whether to treat history cache miss full page reload requests as a "HX-Request" by returning this response header
       * This should always be disabled when using HX-Request header to optionally return partial responses
       * @type boolean
       * @default true
       */
      historyRestoreAsHxRequest: true,
      /**
       * Whether to report input validation errors to the end user and update focus to the first input that fails validation.
       * This should always be enabled as this matches default browser form submit behaviour
       * @type boolean
       * @default false
       */
      reportValidityOfForms: false
    },
    /** @type {typeof parseInterval} */
    parseInterval: null,
    /**
     * proxy of window.location used for page reload functions
     * @type location
     */
    location,
    /** @type {typeof internalEval} */
    _: null,
    version: '2.0.8'
  };
  // Tsc madness part 2
  htmx.onLoad = onLoadHelper;
  htmx.process = processNode;
  htmx.on = addEventListenerImpl;
  htmx.off = removeEventListenerImpl;
  htmx.trigger = triggerEvent;
  htmx.ajax = ajaxHelper;
  htmx.find = find;
  htmx.findAll = findAll;
  htmx.closest = closest;
  htmx.remove = removeElement;
  htmx.addClass = addClassToElement;
  htmx.removeClass = removeClassFromElement;
  htmx.toggleClass = toggleClassOnElement;
  htmx.takeClass = takeClassForElement;
  htmx.swap = swap;
  htmx.defineExtension = defineExtension;
  htmx.removeExtension = removeExtension;
  htmx.logAll = logAll;
  htmx.logNone = logNone;
  htmx.parseInterval = parseInterval;
  htmx._ = internalEval;

  const internalAPI = {
    addTriggerHandler,
    bodyContains,
    canAccessLocalStorage,
    findThisElement,
    filterValues,
    swap,
    hasAttribute,
    getAttributeValue,
    getClosestAttributeValue,
    getClosestMatch,
    getExpressionVars,
    getHeaders,
    getInputValues,
    getInternalData,
    getSwapSpecification,
    getTriggerSpecs,
    getTarget,
    makeFragment,
    mergeObjects,
    makeSettleInfo,
    oobSwap,
    querySelectorExt,
    settleImmediately,
    shouldCancel,
    triggerEvent,
    triggerErrorEvent,
    withExtensions
  };

  const VERBS = ['get', 'post', 'put', 'delete', 'patch'];
  const VERB_SELECTOR = VERBS.map(function(verb) {
    return '[hx-' + verb + '], [data-hx-' + verb + ']'
  }).join(', ');

  //= ===================================================================
  // Utilities
  //= ===================================================================

  /**
   * Parses an interval string consistent with the way htmx does. Useful for plugins that have timing-related attributes.
   *
   * Caution: Accepts an int followed by either **s** or **ms**. All other values use **parseFloat**
   *
   * @see https://htmx.org/api/#parseInterval
   *
   * @param {string} str timing string
   * @returns {number|undefined}
   */
  function parseInterval(str) {
    if (str == undefined) {
      return undefined
    }

    let interval = NaN;
    if (str.slice(-2) == 'ms') {
      interval = parseFloat(str.slice(0, -2));
    } else if (str.slice(-1) == 's') {
      interval = parseFloat(str.slice(0, -1)) * 1000;
    } else if (str.slice(-1) == 'm') {
      interval = parseFloat(str.slice(0, -1)) * 1000 * 60;
    } else {
      interval = parseFloat(str);
    }
    return isNaN(interval) ? undefined : interval
  }

  /**
   * @param {Node} elt
   * @param {string} name
   * @returns {(string | null)}
   */
  function getRawAttribute(elt, name) {
    return elt instanceof Element && elt.getAttribute(name)
  }

  /**
   * @param {Element} elt
   * @param {string} qualifiedName
   * @returns {boolean}
   */
  // resolve with both hx and data-hx prefixes
  function hasAttribute(elt, qualifiedName) {
    return !!elt.hasAttribute && (elt.hasAttribute(qualifiedName) ||
      elt.hasAttribute('data-' + qualifiedName))
  }

  /**
   *
   * @param {Node} elt
   * @param {string} qualifiedName
   * @returns {(string | null)}
   */
  function getAttributeValue(elt, qualifiedName) {
    return getRawAttribute(elt, qualifiedName) || getRawAttribute(elt, 'data-' + qualifiedName)
  }

  /**
   * @param {Node} elt
   * @returns {Node | null}
   */
  function parentElt(elt) {
    const parent = elt.parentElement;
    if (!parent && elt.parentNode instanceof ShadowRoot) return elt.parentNode
    return parent
  }

  /**
   * @returns {Document}
   */
  function getDocument() {
    return document
  }

  /**
   * @param {Node} elt
   * @param {boolean} global
   * @returns {Node|Document}
   */
  function getRootNode(elt, global) {
    return elt.getRootNode ? elt.getRootNode({ composed: global }) : getDocument()
  }

  /**
   * @param {Node} elt
   * @param {(e:Node) => boolean} condition
   * @returns {Node | null}
   */
  function getClosestMatch(elt, condition) {
    while (elt && !condition(elt)) {
      elt = parentElt(elt);
    }

    return elt || null
  }

  /**
   * @param {Element} initialElement
   * @param {Element} ancestor
   * @param {string} attributeName
   * @returns {string|null}
   */
  function getAttributeValueWithDisinheritance(initialElement, ancestor, attributeName) {
    const attributeValue = getAttributeValue(ancestor, attributeName);
    const disinherit = getAttributeValue(ancestor, 'hx-disinherit');
    var inherit = getAttributeValue(ancestor, 'hx-inherit');
    if (initialElement !== ancestor) {
      if (htmx.config.disableInheritance) {
        if (inherit && (inherit === '*' || inherit.split(' ').indexOf(attributeName) >= 0)) {
          return attributeValue
        } else {
          return null
        }
      }
      if (disinherit && (disinherit === '*' || disinherit.split(' ').indexOf(attributeName) >= 0)) {
        return 'unset'
      }
    }
    return attributeValue
  }

  /**
   * @param {Element} elt
   * @param {string} attributeName
   * @returns {string | null}
   */
  function getClosestAttributeValue(elt, attributeName) {
    let closestAttr = null;
    getClosestMatch(elt, function(e) {
      return !!(closestAttr = getAttributeValueWithDisinheritance(elt, asElement(e), attributeName))
    });
    if (closestAttr !== 'unset') {
      return closestAttr
    }
  }

  /**
   * @param {Node} elt
   * @param {string} selector
   * @returns {boolean}
   */
  function matches(elt, selector) {
    return elt instanceof Element && elt.matches(selector)
  }

  /**
   * @param {string} str
   * @returns {string}
   */
  function getStartTag(str) {
    const tagMatcher = /<([a-z][^\/\0>\x20\t\r\n\f]*)/i;
    const match = tagMatcher.exec(str);
    if (match) {
      return match[1].toLowerCase()
    } else {
      return ''
    }
  }

  /**
   * @param {string} resp
   * @returns {Document}
   */
  function parseHTML(resp) {
    if ('parseHTMLUnsafe' in Document) {
      return Document.parseHTMLUnsafe(resp)
    }
    const parser = new DOMParser();
    return parser.parseFromString(resp, 'text/html')
  }

  /**
   * @param {DocumentFragment} fragment
   * @param {Node} elt
   */
  function takeChildrenFor(fragment, elt) {
    while (elt.childNodes.length > 0) {
      fragment.append(elt.childNodes[0]);
    }
  }

  /**
   * @param {HTMLScriptElement} script
   * @returns {HTMLScriptElement}
   */
  function duplicateScript(script) {
    const newScript = getDocument().createElement('script');
    forEach(script.attributes, function(attr) {
      newScript.setAttribute(attr.name, attr.value);
    });
    newScript.textContent = script.textContent;
    newScript.async = false;
    if (htmx.config.inlineScriptNonce) {
      newScript.nonce = htmx.config.inlineScriptNonce;
    }
    return newScript
  }

  /**
   * @param {HTMLScriptElement} script
   * @returns {boolean}
   */
  function isJavaScriptScriptNode(script) {
    return script.matches('script') && (script.type === 'text/javascript' || script.type === 'module' || script.type === '')
  }

  /**
   * we have to make new copies of script tags that we are going to insert because
   * SOME browsers (not saying who, but it involves an element and an animal) don't
   * execute scripts created in <template> tags when they are inserted into the DOM
   * and all the others do lmao
   * @param {DocumentFragment} fragment
   */
  function normalizeScriptTags(fragment) {
    Array.from(fragment.querySelectorAll('script')).forEach(/** @param {HTMLScriptElement} script */ (script) => {
      if (isJavaScriptScriptNode(script)) {
        const newScript = duplicateScript(script);
        const parent = script.parentNode;
        try {
          parent.insertBefore(newScript, script);
        } catch (e) {
          logError(e);
        } finally {
          script.remove();
        }
      }
    });
  }

  /**
   * @typedef {DocumentFragment & {title?: string}} DocumentFragmentWithTitle
   * @description  a document fragment representing the response HTML, including
   * a `title` property for any title information found
   */

  /**
   * @param {string} response HTML
   * @returns {DocumentFragmentWithTitle}
   */
  function makeFragment(response) {
    // strip head tag to determine shape of response we are dealing with
    const responseWithNoHead = response.replace(/<head(\s[^>]*)?>[\s\S]*?<\/head>/i, '');
    const startTag = getStartTag(responseWithNoHead);
    /** @type DocumentFragmentWithTitle */
    let fragment;
    if (startTag === 'html') {
      // if it is a full document, parse it and return the body
      fragment = /** @type DocumentFragmentWithTitle */ (new DocumentFragment());
      const doc = parseHTML(response);
      takeChildrenFor(fragment, doc.body);
      fragment.title = doc.title;
    } else if (startTag === 'body') {
      // parse body w/o wrapping in template
      fragment = /** @type DocumentFragmentWithTitle */ (new DocumentFragment());
      const doc = parseHTML(responseWithNoHead);
      takeChildrenFor(fragment, doc.body);
      fragment.title = doc.title;
    } else {
      // otherwise we have non-body partial HTML content, so wrap it in a template to maximize parsing flexibility
      const doc = parseHTML('<body><template class="internal-htmx-wrapper">' + responseWithNoHead + '</template></body>');
      fragment = /** @type DocumentFragmentWithTitle */ (doc.querySelector('template').content);
      // extract title into fragment for later processing
      fragment.title = doc.title;

      // for legacy reasons we support a title tag at the root level of non-body responses, so we need to handle it
      var titleElement = fragment.querySelector('title');
      if (titleElement && titleElement.parentNode === fragment) {
        titleElement.remove();
        fragment.title = titleElement.innerText;
      }
    }
    if (fragment) {
      if (htmx.config.allowScriptTags) {
        normalizeScriptTags(fragment);
      } else {
        // remove all script tags if scripts are disabled
        fragment.querySelectorAll('script').forEach((script) => script.remove());
      }
    }
    return fragment
  }

  /**
   * @param {Function} func
   */
  function maybeCall(func) {
    if (func) {
      func();
    }
  }

  /**
   * @param {any} o
   * @param {string} type
   * @returns
   */
  function isType(o, type) {
    return Object.prototype.toString.call(o) === '[object ' + type + ']'
  }

  /**
   * @param {*} o
   * @returns {o is Function}
   */
  function isFunction(o) {
    return typeof o === 'function'
  }

  /**
   * @param {*} o
   * @returns {o is Object}
   */
  function isRawObject(o) {
    return isType(o, 'Object')
  }

  /**
   * @typedef {Object} OnHandler
   * @property {(keyof HTMLElementEventMap)|string} event
   * @property {EventListener} listener
   */

  /**
   * @typedef {Object} ListenerInfo
   * @property {string} trigger
   * @property {EventListener} listener
   * @property {EventTarget} on
   */

  /**
   * @typedef {Object} HtmxNodeInternalData
   * Element data
   * @property {number} [initHash]
   * @property {boolean} [boosted]
   * @property {OnHandler[]} [onHandlers]
   * @property {number} [timeout]
   * @property {ListenerInfo[]} [listenerInfos]
   * @property {boolean} [cancelled]
   * @property {boolean} [triggeredOnce]
   * @property {number} [delayed]
   * @property {number|null} [throttle]
   * @property {WeakMap<HtmxTriggerSpecification,WeakMap<EventTarget,string>>} [lastValue]
   * @property {boolean} [loaded]
   * @property {string} [path]
   * @property {string} [verb]
   * @property {boolean} [polling]
   * @property {HTMLButtonElement|HTMLInputElement|null} [lastButtonClicked]
   * @property {number} [requestCount]
   * @property {XMLHttpRequest} [xhr]
   * @property {(() => void)[]} [queuedRequests]
   * @property {boolean} [abortable]
   * @property {boolean} [firstInitCompleted]
   *
   * Event data
   * @property {HtmxTriggerSpecification} [triggerSpec]
   * @property {EventTarget[]} [handledFor]
   */

  /**
   * getInternalData retrieves "private" data stored by htmx within an element
   * @param {EventTarget|Event} elt
   * @returns {HtmxNodeInternalData}
   */
  function getInternalData(elt) {
    const dataProp = 'htmx-internal-data';
    let data = elt[dataProp];
    if (!data) {
      data = elt[dataProp] = {};
    }
    return data
  }

  /**
   * toArray converts an ArrayLike object into a real array.
   * @template T
   * @param {ArrayLike<T>} arr
   * @returns {T[]}
   */
  function toArray(arr) {
    const returnArr = [];
    if (arr) {
      for (let i = 0; i < arr.length; i++) {
        returnArr.push(arr[i]);
      }
    }
    return returnArr
  }

  /**
   * @template T
   * @param {T[]|NamedNodeMap|HTMLCollection|HTMLFormControlsCollection|ArrayLike<T>} arr
   * @param {(T) => void} func
   */
  function forEach(arr, func) {
    if (arr) {
      for (let i = 0; i < arr.length; i++) {
        func(arr[i]);
      }
    }
  }

  /**
   * @param {Element} el
   * @returns {boolean}
   */
  function isScrolledIntoView(el) {
    const rect = el.getBoundingClientRect();
    const elemTop = rect.top;
    const elemBottom = rect.bottom;
    return elemTop < window.innerHeight && elemBottom >= 0
  }

  /**
   * Checks whether the element is in the document (includes shadow roots).
   * This function this is a slight misnomer; it will return true even for elements in the head.
   *
   * @param {Node} elt
   * @returns {boolean}
   */
  function bodyContains(elt) {
    return elt.getRootNode({ composed: true }) === document
  }

  /**
   * @param {string} trigger
   * @returns {string[]}
   */
  function splitOnWhitespace(trigger) {
    return trigger.trim().split(/\s+/)
  }

  /**
   * mergeObjects takes all the keys from
   * obj2 and duplicates them into obj1
   * @template T1
   * @template T2
   * @param {T1} obj1
   * @param {T2} obj2
   * @returns {T1 & T2}
   */
  function mergeObjects(obj1, obj2) {
    for (const key in obj2) {
      if (obj2.hasOwnProperty(key)) {
        // @ts-ignore tsc doesn't seem to properly handle types merging
        obj1[key] = obj2[key];
      }
    }
    // @ts-ignore tsc doesn't seem to properly handle types merging
    return obj1
  }

  /**
   * @param {string} jString
   * @returns {any|null}
   */
  function parseJSON(jString) {
    try {
      return JSON.parse(jString)
    } catch (error) {
      logError(error);
      return null
    }
  }

  /**
   * @returns {boolean}
   */
  function canAccessLocalStorage() {
    const test = 'htmx:sessionStorageTest';
    try {
      sessionStorage.setItem(test, test);
      sessionStorage.removeItem(test);
      return true
    } catch (e) {
      return false
    }
  }

  /**
   * @param {string} path
   * @returns {string}
   */
  function normalizePath(path) {
    // use dummy base URL to allow normalize on path only
    const url = new URL(path, 'http://x');
    if (url) {
      path = url.pathname + url.search;
    }
    // remove trailing slash, unless index page
    if (path != '/') {
      path = path.replace(/\/+$/, '');
    }
    return path
  }

  //= =========================================================================================
  // public API
  //= =========================================================================================

  /**
   * @param {string} str
   * @returns {any}
   */
  function internalEval(str) {
    return maybeEval(getDocument().body, function() {
      return eval(str)
    })
  }

  /**
   * Adds a callback for the **htmx:load** event. This can be used to process new content, for example initializing the content with a javascript library
   *
   * @see https://htmx.org/api/#onLoad
   *
   * @param {(elt: Node) => void} callback the callback to call on newly loaded content
   * @returns {EventListener}
   */
  function onLoadHelper(callback) {
    const value = htmx.on('htmx:load', /** @param {CustomEvent} evt */ function(evt) {
      callback(evt.detail.elt);
    });
    return value
  }

  /**
   * Log all htmx events, useful for debugging.
   *
   * @see https://htmx.org/api/#logAll
   */
  function logAll() {
    htmx.logger = function(elt, event, data) {
      if (console) {
        console.log(event, elt, data);
      }
    };
  }

  function logNone() {
    htmx.logger = null;
  }

  /**
   * Finds an element matching the selector
   *
   * @see https://htmx.org/api/#find
   *
   * @param {ParentNode|string} eltOrSelector  the root element to find the matching element in, inclusive | the selector to match
   * @param {string} [selector] the selector to match
   * @returns {Element|null}
   */
  function find(eltOrSelector, selector) {
    if (typeof eltOrSelector !== 'string') {
      return eltOrSelector.querySelector(selector)
    } else {
      return find(getDocument(), eltOrSelector)
    }
  }

  /**
   * Finds all elements matching the selector
   *
   * @see https://htmx.org/api/#findAll
   *
   * @param {ParentNode|string} eltOrSelector the root element to find the matching elements in, inclusive | the selector to match
   * @param {string} [selector] the selector to match
   * @returns {NodeListOf<Element>}
   */
  function findAll(eltOrSelector, selector) {
    if (typeof eltOrSelector !== 'string') {
      return eltOrSelector.querySelectorAll(selector)
    } else {
      return findAll(getDocument(), eltOrSelector)
    }
  }

  /**
   * @returns Window
   */
  function getWindow() {
    return window
  }

  /**
   * Removes an element from the DOM
   *
   * @see https://htmx.org/api/#remove
   *
   * @param {Node} elt
   * @param {number} [delay]
   */
  function removeElement(elt, delay) {
    elt = resolveTarget(elt);
    if (delay) {
      getWindow().setTimeout(function() {
        removeElement(elt);
        elt = null;
      }, delay);
    } else {
      parentElt(elt).removeChild(elt);
    }
  }

  /**
   * @param {any} elt
   * @return {Element|null}
   */
  function asElement(elt) {
    return elt instanceof Element ? elt : null
  }

  /**
   * @param {any} elt
   * @return {HTMLElement|null}
   */
  function asHtmlElement(elt) {
    return elt instanceof HTMLElement ? elt : null
  }

  /**
   * @param {any} value
   * @return {string|null}
   */
  function asString(value) {
    return typeof value === 'string' ? value : null
  }

  /**
   * @param {EventTarget} elt
   * @return {ParentNode|null}
   */
  function asParentNode(elt) {
    return elt instanceof Element || elt instanceof Document || elt instanceof DocumentFragment ? elt : null
  }

  /**
   * This method adds a class to the given element.
   *
   * @see https://htmx.org/api/#addClass
   *
   * @param {Element|string} elt the element to add the class to
   * @param {string} clazz the class to add
   * @param {number} [delay] the delay (in milliseconds) before class is added
   */
  function addClassToElement(elt, clazz, delay) {
    elt = asElement(resolveTarget(elt));
    if (!elt) {
      return
    }
    if (delay) {
      getWindow().setTimeout(function() {
        addClassToElement(elt, clazz);
        elt = null;
      }, delay);
    } else {
      elt.classList && elt.classList.add(clazz);
    }
  }

  /**
   * Removes a class from the given element
   *
   * @see https://htmx.org/api/#removeClass
   *
   * @param {Node|string} node element to remove the class from
   * @param {string} clazz the class to remove
   * @param {number} [delay] the delay (in milliseconds before class is removed)
   */
  function removeClassFromElement(node, clazz, delay) {
    let elt = asElement(resolveTarget(node));
    if (!elt) {
      return
    }
    if (delay) {
      getWindow().setTimeout(function() {
        removeClassFromElement(elt, clazz);
        elt = null;
      }, delay);
    } else {
      if (elt.classList) {
        elt.classList.remove(clazz);
        // if there are no classes left, remove the class attribute
        if (elt.classList.length === 0) {
          elt.removeAttribute('class');
        }
      }
    }
  }

  /**
   * Toggles the given class on an element
   *
   * @see https://htmx.org/api/#toggleClass
   *
   * @param {Element|string} elt the element to toggle the class on
   * @param {string} clazz the class to toggle
   */
  function toggleClassOnElement(elt, clazz) {
    elt = resolveTarget(elt);
    elt.classList.toggle(clazz);
  }

  /**
   * Takes the given class from its siblings, so that among its siblings, only the given element will have the class.
   *
   * @see https://htmx.org/api/#takeClass
   *
   * @param {Node|string} elt the element that will take the class
   * @param {string} clazz the class to take
   */
  function takeClassForElement(elt, clazz) {
    elt = resolveTarget(elt);
    forEach(elt.parentElement.children, function(child) {
      removeClassFromElement(child, clazz);
    });
    addClassToElement(asElement(elt), clazz);
  }

  /**
   * Finds the closest matching element in the given elements parentage, inclusive of the element
   *
   * @see https://htmx.org/api/#closest
   *
   * @param {Element|string} elt the element to find the selector from
   * @param {string} selector the selector to find
   * @returns {Element|null}
   */
  function closest(elt, selector) {
    elt = asElement(resolveTarget(elt));
    if (elt) {
      return elt.closest(selector)
    }
    return null
  }

  /**
   * @param {string} str
   * @param {string} prefix
   * @returns {boolean}
   */
  function startsWith(str, prefix) {
    return str.substring(0, prefix.length) === prefix
  }

  /**
   * @param {string} str
   * @param {string} suffix
   * @returns {boolean}
   */
  function endsWith(str, suffix) {
    return str.substring(str.length - suffix.length) === suffix
  }

  /**
   * @param {string} selector
   * @returns {string}
   */
  function normalizeSelector(selector) {
    const trimmedSelector = selector.trim();
    if (startsWith(trimmedSelector, '<') && endsWith(trimmedSelector, '/>')) {
      return trimmedSelector.substring(1, trimmedSelector.length - 2)
    } else {
      return trimmedSelector
    }
  }

  /**
   * @param {Node|Element|Document|string} elt
   * @param {string} selector
   * @param {boolean=} global
   * @returns {(Node|Window)[]}
   */
  function querySelectorAllExt(elt, selector, global) {
    if (selector.indexOf('global ') === 0) {
      return querySelectorAllExt(elt, selector.slice(7), true)
    }

    elt = resolveTarget(elt);

    const parts = [];
    {
      let chevronsCount = 0;
      let offset = 0;
      for (let i = 0; i < selector.length; i++) {
        const char = selector[i];
        if (char === ',' && chevronsCount === 0) {
          parts.push(selector.substring(offset, i));
          offset = i + 1;
          continue
        }
        if (char === '<') {
          chevronsCount++;
        } else if (char === '/' && i < selector.length - 1 && selector[i + 1] === '>') {
          chevronsCount--;
        }
      }
      if (offset < selector.length) {
        parts.push(selector.substring(offset));
      }
    }

    const result = [];
    const unprocessedParts = [];
    while (parts.length > 0) {
      const selector = normalizeSelector(parts.shift());
      let item;
      if (selector.indexOf('closest ') === 0) {
        item = closest(asElement(elt), normalizeSelector(selector.slice(8)));
      } else if (selector.indexOf('find ') === 0) {
        item = find(asParentNode(elt), normalizeSelector(selector.slice(5)));
      } else if (selector === 'next' || selector === 'nextElementSibling') {
        item = asElement(elt).nextElementSibling;
      } else if (selector.indexOf('next ') === 0) {
        item = scanForwardQuery(elt, normalizeSelector(selector.slice(5)), !!global);
      } else if (selector === 'previous' || selector === 'previousElementSibling') {
        item = asElement(elt).previousElementSibling;
      } else if (selector.indexOf('previous ') === 0) {
        item = scanBackwardsQuery(elt, normalizeSelector(selector.slice(9)), !!global);
      } else if (selector === 'document') {
        item = document;
      } else if (selector === 'window') {
        item = window;
      } else if (selector === 'body') {
        item = document.body;
      } else if (selector === 'root') {
        item = getRootNode(elt, !!global);
      } else if (selector === 'host') {
        item = (/** @type ShadowRoot */(elt.getRootNode())).host;
      } else {
        unprocessedParts.push(selector);
      }

      if (item) {
        result.push(item);
      }
    }

    if (unprocessedParts.length > 0) {
      const standardSelector = unprocessedParts.join(',');
      const rootNode = asParentNode(getRootNode(elt, !!global));
      result.push(...toArray(rootNode.querySelectorAll(standardSelector)));
    }

    return result
  }

  /**
   * @param {Node} start
   * @param {string} match
   * @param {boolean} global
   * @returns {Element}
   */
  var scanForwardQuery = function(start, match, global) {
    const results = asParentNode(getRootNode(start, global)).querySelectorAll(match);
    for (let i = 0; i < results.length; i++) {
      const elt = results[i];
      if (elt.compareDocumentPosition(start) === Node.DOCUMENT_POSITION_PRECEDING) {
        return elt
      }
    }
  };

  /**
   * @param {Node} start
   * @param {string} match
   * @param {boolean} global
   * @returns {Element}
   */
  var scanBackwardsQuery = function(start, match, global) {
    const results = asParentNode(getRootNode(start, global)).querySelectorAll(match);
    for (let i = results.length - 1; i >= 0; i--) {
      const elt = results[i];
      if (elt.compareDocumentPosition(start) === Node.DOCUMENT_POSITION_FOLLOWING) {
        return elt
      }
    }
  };

  /**
   * @param {Node|string} eltOrSelector
   * @param {string=} selector
   * @returns {Node|Window}
   */
  function querySelectorExt(eltOrSelector, selector) {
    if (typeof eltOrSelector !== 'string') {
      return querySelectorAllExt(eltOrSelector, selector)[0]
    } else {
      return querySelectorAllExt(getDocument().body, eltOrSelector)[0]
    }
  }

  /**
   * @template {EventTarget} T
   * @param {T|string} eltOrSelector
   * @param {T} [context]
   * @returns {Element|T|null}
   */
  function resolveTarget(eltOrSelector, context) {
    if (typeof eltOrSelector === 'string') {
      return find(asParentNode(context) || document, eltOrSelector)
    } else {
      return eltOrSelector
    }
  }

  /**
   * @typedef {keyof HTMLElementEventMap|string} AnyEventName
   */

  /**
   * @typedef {Object} EventArgs
   * @property {EventTarget} target
   * @property {AnyEventName} event
   * @property {EventListener} listener
   * @property {Object|boolean} options
   */

  /**
   * @param {EventTarget|AnyEventName} arg1
   * @param {AnyEventName|EventListener} arg2
   * @param {EventListener|Object|boolean} [arg3]
   * @param {Object|boolean} [arg4]
   * @returns {EventArgs}
   */
  function processEventArgs(arg1, arg2, arg3, arg4) {
    if (isFunction(arg2)) {
      return {
        target: getDocument().body,
        event: asString(arg1),
        listener: arg2,
        options: arg3
      }
    } else {
      return {
        target: resolveTarget(arg1),
        event: asString(arg2),
        listener: arg3,
        options: arg4
      }
    }
  }

  /**
   * Adds an event listener to an element
   *
   * @see https://htmx.org/api/#on
   *
   * @param {EventTarget|string} arg1 the element to add the listener to | the event name to add the listener for
   * @param {string|EventListener} arg2 the event name to add the listener for | the listener to add
   * @param {EventListener|Object|boolean} [arg3] the listener to add | options to add
   * @param {Object|boolean} [arg4] options to add
   * @returns {EventListener}
   */
  function addEventListenerImpl(arg1, arg2, arg3, arg4) {
    ready(function() {
      const eventArgs = processEventArgs(arg1, arg2, arg3, arg4);
      eventArgs.target.addEventListener(eventArgs.event, eventArgs.listener, eventArgs.options);
    });
    const b = isFunction(arg2);
    return b ? arg2 : arg3
  }

  /**
   * Removes an event listener from an element
   *
   * @see https://htmx.org/api/#off
   *
   * @param {EventTarget|string} arg1 the element to remove the listener from | the event name to remove the listener from
   * @param {string|EventListener} arg2 the event name to remove the listener from | the listener to remove
   * @param {EventListener} [arg3] the listener to remove
   * @returns {EventListener}
   */
  function removeEventListenerImpl(arg1, arg2, arg3) {
    ready(function() {
      const eventArgs = processEventArgs(arg1, arg2, arg3);
      eventArgs.target.removeEventListener(eventArgs.event, eventArgs.listener);
    });
    return isFunction(arg2) ? arg2 : arg3
  }

  //= ===================================================================
  // Node processing
  //= ===================================================================

  const DUMMY_ELT = getDocument().createElement('output'); // dummy element for bad selectors
  /**
   * @param {Element} elt
   * @param {string} attrName
   * @returns {(Node|Window)[]}
   */
  function findAttributeTargets(elt, attrName) {
    const attrTarget = getClosestAttributeValue(elt, attrName);
    if (attrTarget) {
      if (attrTarget === 'this') {
        return [findThisElement(elt, attrName)]
      } else {
        const result = querySelectorAllExt(elt, attrTarget);
        // find `inherit` whole word in value, make sure it's surrounded by commas or is at the start/end of string
        const shouldInherit = /(^|,)(\s*)inherit(\s*)($|,)/.test(attrTarget);
        if (shouldInherit) {
          const eltToInheritFrom = asElement(getClosestMatch(elt, function(parent) {
            return parent !== elt && hasAttribute(asElement(parent), attrName)
          }));
          if (eltToInheritFrom) {
            result.push(...findAttributeTargets(eltToInheritFrom, attrName));
          }
        }
        if (result.length === 0) {
          logError('The selector "' + attrTarget + '" on ' + attrName + ' returned no matches!');
          return [DUMMY_ELT]
        } else {
          return result
        }
      }
    }
  }

  /**
   * @param {Element} elt
   * @param {string} attribute
   * @returns {Element|null}
   */
  function findThisElement(elt, attribute) {
    return asElement(getClosestMatch(elt, function(elt) {
      return getAttributeValue(asElement(elt), attribute) != null
    }))
  }

  /**
   * @param {Element} elt
   * @returns {Node|Window|null}
   */
  function getTarget(elt) {
    const targetStr = getClosestAttributeValue(elt, 'hx-target');
    if (targetStr) {
      if (targetStr === 'this') {
        return findThisElement(elt, 'hx-target')
      } else {
        return querySelectorExt(elt, targetStr)
      }
    } else {
      const data = getInternalData(elt);
      if (data.boosted) {
        return getDocument().body
      } else {
        return elt
      }
    }
  }

  /**
   * @param {string} name
   * @returns {boolean}
   */
  function shouldSettleAttribute(name) {
    return htmx.config.attributesToSettle.includes(name)
  }

  /**
   * @param {Element} mergeTo
   * @param {Element} mergeFrom
   */
  function cloneAttributes(mergeTo, mergeFrom) {
    forEach(Array.from(mergeTo.attributes), function(attr) {
      if (!mergeFrom.hasAttribute(attr.name) && shouldSettleAttribute(attr.name)) {
        mergeTo.removeAttribute(attr.name);
      }
    });
    forEach(mergeFrom.attributes, function(attr) {
      if (shouldSettleAttribute(attr.name)) {
        mergeTo.setAttribute(attr.name, attr.value);
      }
    });
  }

  /**
   * @param {HtmxSwapStyle} swapStyle
   * @param {Element} target
   * @returns {boolean}
   */
  function isInlineSwap(swapStyle, target) {
    const extensions = getExtensions(target);
    for (let i = 0; i < extensions.length; i++) {
      const extension = extensions[i];
      try {
        if (extension.isInlineSwap(swapStyle)) {
          return true
        }
      } catch (e) {
        logError(e);
      }
    }
    return swapStyle === 'outerHTML'
  }

  /**
   * @param {string} oobValue
   * @param {Element} oobElement
   * @param {HtmxSettleInfo} settleInfo
   * @param {Node|Document} [rootNode]
   * @returns
   */
  function oobSwap(oobValue, oobElement, settleInfo, rootNode) {
    rootNode = rootNode || getDocument();
    let selector = '#' + CSS.escape(getRawAttribute(oobElement, 'id'));
    /** @type HtmxSwapStyle */
    let swapStyle = 'outerHTML';
    if (oobValue === 'true') ; else if (oobValue.indexOf(':') > 0) {
      swapStyle = oobValue.substring(0, oobValue.indexOf(':'));
      selector = oobValue.substring(oobValue.indexOf(':') + 1);
    } else {
      swapStyle = oobValue;
    }
    oobElement.removeAttribute('hx-swap-oob');
    oobElement.removeAttribute('data-hx-swap-oob');

    const targets = querySelectorAllExt(rootNode, selector, false);
    if (targets.length) {
      forEach(
        targets,
        function(target) {
          let fragment;
          const oobElementClone = oobElement.cloneNode(true);
          fragment = getDocument().createDocumentFragment();
          fragment.appendChild(oobElementClone);
          if (!isInlineSwap(swapStyle, target)) {
            fragment = asParentNode(oobElementClone); // if this is not an inline swap, we use the content of the node, not the node itself
          }

          const beforeSwapDetails = { shouldSwap: true, target, fragment };
          if (!triggerEvent(target, 'htmx:oobBeforeSwap', beforeSwapDetails)) return

          target = beforeSwapDetails.target; // allow re-targeting
          if (beforeSwapDetails.shouldSwap) {
            handlePreservedElements(fragment);
            swapWithStyle(swapStyle, target, target, fragment, settleInfo);
            restorePreservedElements();
          }
          forEach(settleInfo.elts, function(elt) {
            triggerEvent(elt, 'htmx:oobAfterSwap', beforeSwapDetails);
          });
        }
      );
      oobElement.parentNode.removeChild(oobElement);
    } else {
      oobElement.parentNode.removeChild(oobElement);
      triggerErrorEvent(getDocument().body, 'htmx:oobErrorNoTarget', { content: oobElement });
    }
    return oobValue
  }

  function restorePreservedElements() {
    const pantry = find('#--htmx-preserve-pantry--');
    if (pantry) {
      for (const preservedElt of [...pantry.children]) {
        const existingElement = find('#' + preservedElt.id);
        // @ts-ignore - use proposed moveBefore feature
        existingElement.parentNode.moveBefore(preservedElt, existingElement);
        existingElement.remove();
      }
      pantry.remove();
    }
  }

  /**
   * @param {DocumentFragment|ParentNode} fragment
   */
  function handlePreservedElements(fragment) {
    forEach(findAll(fragment, '[hx-preserve], [data-hx-preserve]'), function(preservedElt) {
      const id = getAttributeValue(preservedElt, 'id');
      const existingElement = getDocument().getElementById(id);
      if (existingElement != null) {
        if (preservedElt.moveBefore) { // if the moveBefore API exists, use it
          // get or create a storage spot for stuff
          let pantry = find('#--htmx-preserve-pantry--');
          if (pantry == null) {
            getDocument().body.insertAdjacentHTML('afterend', "<div id='--htmx-preserve-pantry--'></div>");
            pantry = find('#--htmx-preserve-pantry--');
          }
          // @ts-ignore - use proposed moveBefore feature
          pantry.moveBefore(existingElement, null);
        } else {
          preservedElt.parentNode.replaceChild(existingElement, preservedElt);
        }
      }
    });
  }

  /**
   * @param {Node} parentNode
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function handleAttributes(parentNode, fragment, settleInfo) {
    forEach(fragment.querySelectorAll('[id]'), function(newNode) {
      const id = getRawAttribute(newNode, 'id');
      if (id && id.length > 0) {
        const normalizedId = id.replace("'", "\\'");
        const normalizedTag = newNode.tagName.replace(':', '\\:');
        const parentElt = asParentNode(parentNode);
        const oldNode = parentElt && parentElt.querySelector(normalizedTag + "[id='" + normalizedId + "']");
        if (oldNode && oldNode !== parentElt) {
          const newAttributes = newNode.cloneNode();
          cloneAttributes(newNode, oldNode);
          settleInfo.tasks.push(function() {
            cloneAttributes(newNode, newAttributes);
          });
        }
      }
    });
  }

  /**
   * @param {Node} child
   * @returns {HtmxSettleTask}
   */
  function makeAjaxLoadTask(child) {
    return function() {
      removeClassFromElement(child, htmx.config.addedClass);
      processNode(asElement(child));
      processFocus(asParentNode(child));
      triggerEvent(child, 'htmx:load');
    }
  }

  /**
   * @param {ParentNode} child
   */
  function processFocus(child) {
    const autofocus = '[autofocus]';
    const autoFocusedElt = asHtmlElement(matches(child, autofocus) ? child : child.querySelector(autofocus));
    if (autoFocusedElt != null) {
      autoFocusedElt.focus();
    }
  }

  /**
   * @param {Node} parentNode
   * @param {Node} insertBefore
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function insertNodesBefore(parentNode, insertBefore, fragment, settleInfo) {
    handleAttributes(parentNode, fragment, settleInfo);
    while (fragment.childNodes.length > 0) {
      const child = fragment.firstChild;
      addClassToElement(asElement(child), htmx.config.addedClass);
      parentNode.insertBefore(child, insertBefore);
      if (child.nodeType !== Node.TEXT_NODE && child.nodeType !== Node.COMMENT_NODE) {
        settleInfo.tasks.push(makeAjaxLoadTask(child));
      }
    }
  }

  /**
   * based on https://gist.github.com/hyamamoto/fd435505d29ebfa3d9716fd2be8d42f0,
   * derived from Java's string hashcode implementation
   * @param {string} string
   * @param {number} hash
   * @returns {number}
   */
  function stringHash(string, hash) {
    let char = 0;
    while (char < string.length) {
      hash = (hash << 5) - hash + string.charCodeAt(char++) | 0; // bitwise or ensures we have a 32-bit int
    }
    return hash
  }

  /**
   * @param {Element} elt
   * @returns {number}
   */
  function attributeHash(elt) {
    let hash = 0;
    for (let i = 0; i < elt.attributes.length; i++) {
      const attribute = elt.attributes[i];
      if (attribute.value) { // only include attributes w/ actual values (empty is same as non-existent)
        hash = stringHash(attribute.name, hash);
        hash = stringHash(attribute.value, hash);
      }
    }
    return hash
  }

  /**
   * @param {EventTarget} elt
   */
  function deInitOnHandlers(elt) {
    const internalData = getInternalData(elt);
    if (internalData.onHandlers) {
      for (let i = 0; i < internalData.onHandlers.length; i++) {
        const handlerInfo = internalData.onHandlers[i];
        removeEventListenerImpl(elt, handlerInfo.event, handlerInfo.listener);
      }
      delete internalData.onHandlers;
    }
  }

  /**
   * @param {Node} element
   */
  function deInitNode(element) {
    const internalData = getInternalData(element);
    if (internalData.timeout) {
      clearTimeout(internalData.timeout);
    }
    if (internalData.listenerInfos) {
      forEach(internalData.listenerInfos, function(info) {
        if (info.on) {
          removeEventListenerImpl(info.on, info.trigger, info.listener);
        }
      });
    }
    deInitOnHandlers(element);
    forEach(Object.keys(internalData), function(key) { if (key !== 'firstInitCompleted') delete internalData[key]; });
  }

  /**
   * @param {Node} element
   */
  function cleanUpElement(element) {
    triggerEvent(element, 'htmx:beforeCleanupElement');
    deInitNode(element);
    // @ts-ignore
    forEach(element.children, function(child) { cleanUpElement(child); });
  }

  /**
   * @param {Element} target
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function swapOuterHTML(target, fragment, settleInfo) {
    if (target.tagName === 'BODY') { // special case the body to innerHTML because DocumentFragments can't contain a body elt unfortunately
      return swapInnerHTML(target, fragment, settleInfo)
    }
    /** @type {Node} */
    let newElt;
    const eltBeforeNewContent = target.previousSibling;
    const parentNode = parentElt(target);
    if (!parentNode) { // when parent node disappears, we can't do anything
      return
    }
    insertNodesBefore(parentNode, target, fragment, settleInfo);
    if (eltBeforeNewContent == null) {
      newElt = parentNode.firstChild;
    } else {
      newElt = eltBeforeNewContent.nextSibling;
    }
    settleInfo.elts = settleInfo.elts.filter(function(e) { return e !== target });
    // scan through all newly added content and add all elements to the settle info so we trigger
    // events properly on them
    while (newElt && newElt !== target) {
      if (newElt instanceof Element) {
        settleInfo.elts.push(newElt);
      }
      newElt = newElt.nextSibling;
    }
    cleanUpElement(target);
    target.remove();
  }

  /**
   * @param {Element} target
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function swapAfterBegin(target, fragment, settleInfo) {
    return insertNodesBefore(target, target.firstChild, fragment, settleInfo)
  }

  /**
   * @param {Element} target
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function swapBeforeBegin(target, fragment, settleInfo) {
    return insertNodesBefore(parentElt(target), target, fragment, settleInfo)
  }

  /**
   * @param {Element} target
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function swapBeforeEnd(target, fragment, settleInfo) {
    return insertNodesBefore(target, null, fragment, settleInfo)
  }

  /**
   * @param {Element} target
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function swapAfterEnd(target, fragment, settleInfo) {
    return insertNodesBefore(parentElt(target), target.nextSibling, fragment, settleInfo)
  }

  /**
   * @param {Element} target
   */
  function swapDelete(target) {
    cleanUpElement(target);
    const parent = parentElt(target);
    if (parent) {
      return parent.removeChild(target)
    }
  }

  /**
   * @param {Element} target
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function swapInnerHTML(target, fragment, settleInfo) {
    const firstChild = target.firstChild;
    insertNodesBefore(target, firstChild, fragment, settleInfo);
    if (firstChild) {
      while (firstChild.nextSibling) {
        cleanUpElement(firstChild.nextSibling);
        target.removeChild(firstChild.nextSibling);
      }
      cleanUpElement(firstChild);
      target.removeChild(firstChild);
    }
  }

  /**
   * @param {HtmxSwapStyle} swapStyle
   * @param {Element} elt
   * @param {Element} target
   * @param {ParentNode} fragment
   * @param {HtmxSettleInfo} settleInfo
   */
  function swapWithStyle(swapStyle, elt, target, fragment, settleInfo) {
    switch (swapStyle) {
      case 'none':
        return
      case 'outerHTML':
        swapOuterHTML(target, fragment, settleInfo);
        return
      case 'afterbegin':
        swapAfterBegin(target, fragment, settleInfo);
        return
      case 'beforebegin':
        swapBeforeBegin(target, fragment, settleInfo);
        return
      case 'beforeend':
        swapBeforeEnd(target, fragment, settleInfo);
        return
      case 'afterend':
        swapAfterEnd(target, fragment, settleInfo);
        return
      case 'delete':
        swapDelete(target);
        return
      default:
        var extensions = getExtensions(elt);
        for (let i = 0; i < extensions.length; i++) {
          const ext = extensions[i];
          try {
            const newElements = ext.handleSwap(swapStyle, target, fragment, settleInfo);
            if (newElements) {
              if (Array.isArray(newElements)) {
                // if handleSwap returns an array (like) of elements, we handle them
                for (let j = 0; j < newElements.length; j++) {
                  const child = newElements[j];
                  if (child.nodeType !== Node.TEXT_NODE && child.nodeType !== Node.COMMENT_NODE) {
                    settleInfo.tasks.push(makeAjaxLoadTask(child));
                  }
                }
              }
              return
            }
          } catch (e) {
            logError(e);
          }
        }
        if (swapStyle === 'innerHTML') {
          swapInnerHTML(target, fragment, settleInfo);
        } else {
          swapWithStyle(htmx.config.defaultSwapStyle, elt, target, fragment, settleInfo);
        }
    }
  }

  /**
   * @param {DocumentFragment} fragment
   * @param {HtmxSettleInfo} settleInfo
   * @param {Node|Document} [rootNode]
   */
  function findAndSwapOobElements(fragment, settleInfo, rootNode) {
    var oobElts = findAll(fragment, '[hx-swap-oob], [data-hx-swap-oob]');
    forEach(oobElts, function(oobElement) {
      if (htmx.config.allowNestedOobSwaps || oobElement.parentElement === null) {
        const oobValue = getAttributeValue(oobElement, 'hx-swap-oob');
        if (oobValue != null) {
          oobSwap(oobValue, oobElement, settleInfo, rootNode);
        }
      } else {
        oobElement.removeAttribute('hx-swap-oob');
        oobElement.removeAttribute('data-hx-swap-oob');
      }
    });
    return oobElts.length > 0
  }

  /**
   * Implements complete swapping pipeline, including: delay, view transitions, focus and selection preservation,
   * title updates, scroll, OOB swapping, normal swapping and settling
   * @param {string|Element} target
   * @param {string} content
   * @param {HtmxSwapSpecification} swapSpec
   * @param {SwapOptions} [swapOptions]
   */
  function swap(target, content, swapSpec, swapOptions) {
    if (!swapOptions) {
      swapOptions = {};
    }
    // optional transition API promise callbacks
    let settleResolve = null;
    let settleReject = null;

    let doSwap = function() {
      maybeCall(swapOptions.beforeSwapCallback);

      target = resolveTarget(target);
      const rootNode = swapOptions.contextElement ? getRootNode(swapOptions.contextElement, false) : getDocument();

      // preserve focus and selection
      const activeElt = document.activeElement;
      let selectionInfo = {};
      selectionInfo = {
        elt: activeElt,
        // @ts-ignore
        start: activeElt ? activeElt.selectionStart : null,
        // @ts-ignore
        end: activeElt ? activeElt.selectionEnd : null
      };
      const settleInfo = makeSettleInfo(target);

      // For text content swaps, don't parse the response as HTML, just insert it
      if (swapSpec.swapStyle === 'textContent') {
        target.textContent = content;
      // Otherwise, make the fragment and process it
      } else {
        let fragment = makeFragment(content);

        settleInfo.title = swapOptions.title || fragment.title;
        if (swapOptions.historyRequest) {
          // @ts-ignore fragment can be a parentNode Element
          fragment = fragment.querySelector('[hx-history-elt],[data-hx-history-elt]') || fragment;
        }

        // select-oob swaps
        if (swapOptions.selectOOB) {
          const oobSelectValues = swapOptions.selectOOB.split(',');
          for (let i = 0; i < oobSelectValues.length; i++) {
            const oobSelectValue = oobSelectValues[i].split(':', 2);
            let id = oobSelectValue[0].trim();
            if (id.indexOf('#') === 0) {
              id = id.substring(1);
            }
            const oobValue = oobSelectValue[1] || 'true';
            const oobElement = fragment.querySelector('#' + id);
            if (oobElement) {
              oobSwap(oobValue, oobElement, settleInfo, rootNode);
            }
          }
        }
        // oob swaps
        findAndSwapOobElements(fragment, settleInfo, rootNode);
        forEach(findAll(fragment, 'template'), /** @param {HTMLTemplateElement} template */function(template) {
          if (template.content && findAndSwapOobElements(template.content, settleInfo, rootNode)) {
            // Avoid polluting the DOM with empty templates that were only used to encapsulate oob swap
            template.remove();
          }
        });

        // normal swap
        if (swapOptions.select) {
          const newFragment = getDocument().createDocumentFragment();
          forEach(fragment.querySelectorAll(swapOptions.select), function(node) {
            newFragment.appendChild(node);
          });
          fragment = newFragment;
        }
        handlePreservedElements(fragment);
        swapWithStyle(swapSpec.swapStyle, swapOptions.contextElement, target, fragment, settleInfo);
        restorePreservedElements();
      }

      // apply saved focus and selection information to swapped content
      if (selectionInfo.elt &&
        !bodyContains(selectionInfo.elt) &&
        getRawAttribute(selectionInfo.elt, 'id')) {
        const newActiveElt = document.getElementById(getRawAttribute(selectionInfo.elt, 'id'));
        const focusOptions = { preventScroll: swapSpec.focusScroll !== undefined ? !swapSpec.focusScroll : !htmx.config.defaultFocusScroll };
        if (newActiveElt) {
          // @ts-ignore
          if (selectionInfo.start && newActiveElt.setSelectionRange) {
            try {
              // @ts-ignore
              newActiveElt.setSelectionRange(selectionInfo.start, selectionInfo.end);
            } catch (e) {
              // the setSelectionRange method is present on fields that don't support it, so just let this fail
            }
          }
          newActiveElt.focus(focusOptions);
        }
      }

      target.classList.remove(htmx.config.swappingClass);
      forEach(settleInfo.elts, function(elt) {
        if (elt.classList) {
          elt.classList.add(htmx.config.settlingClass);
        }
        triggerEvent(elt, 'htmx:afterSwap', swapOptions.eventInfo);
      });
      maybeCall(swapOptions.afterSwapCallback);

      // merge in new title after swap but before settle
      if (!swapSpec.ignoreTitle) {
        handleTitle(settleInfo.title);
      }

      // settle
      const doSettle = function() {
        forEach(settleInfo.tasks, function(task) {
          task.call();
        });
        forEach(settleInfo.elts, function(elt) {
          if (elt.classList) {
            elt.classList.remove(htmx.config.settlingClass);
          }
          triggerEvent(elt, 'htmx:afterSettle', swapOptions.eventInfo);
        });

        if (swapOptions.anchor) {
          const anchorTarget = asElement(resolveTarget('#' + swapOptions.anchor));
          if (anchorTarget) {
            anchorTarget.scrollIntoView({ block: 'start', behavior: 'auto' });
          }
        }

        updateScrollState(settleInfo.elts, swapSpec);
        maybeCall(swapOptions.afterSettleCallback);
        maybeCall(settleResolve);
      };

      if (swapSpec.settleDelay > 0) {
        getWindow().setTimeout(doSettle, swapSpec.settleDelay);
      } else {
        doSettle();
      }
    };
    let shouldTransition = htmx.config.globalViewTransitions;
    if (swapSpec.hasOwnProperty('transition')) {
      shouldTransition = swapSpec.transition;
    }

    const elt = swapOptions.contextElement || getDocument();

    if (shouldTransition &&
            triggerEvent(elt, 'htmx:beforeTransition', swapOptions.eventInfo) &&
            typeof Promise !== 'undefined' &&
            // @ts-ignore experimental feature atm
            document.startViewTransition) {
      const settlePromise = new Promise(function(_resolve, _reject) {
        settleResolve = _resolve;
        settleReject = _reject;
      });
      // wrap the original doSwap() in a call to startViewTransition()
      const innerDoSwap = doSwap;
      doSwap = function() {
        // @ts-ignore experimental feature atm
        document.startViewTransition(function() {
          innerDoSwap();
          return settlePromise
        });
      };
    }

    try {
      if (swapSpec?.swapDelay && swapSpec.swapDelay > 0) {
        getWindow().setTimeout(doSwap, swapSpec.swapDelay);
      } else {
        doSwap();
      }
    } catch (e) {
      triggerErrorEvent(elt, 'htmx:swapError', swapOptions.eventInfo);
      maybeCall(settleReject);
      throw e
    }
  }

  /**
   * @param {XMLHttpRequest} xhr
   * @param {string} header
   * @param {EventTarget} elt
   */
  function handleTriggerHeader(xhr, header, elt) {
    const triggerBody = xhr.getResponseHeader(header);
    if (triggerBody.indexOf('{') === 0) {
      const triggers = parseJSON(triggerBody);
      for (const eventName in triggers) {
        if (triggers.hasOwnProperty(eventName)) {
          let detail = triggers[eventName];
          if (isRawObject(detail)) {
            // @ts-ignore
            elt = detail.target !== undefined ? detail.target : elt;
          } else {
            detail = { value: detail };
          }
          triggerEvent(elt, eventName, detail);
        }
      }
    } else {
      const eventNames = triggerBody.split(',');
      for (let i = 0; i < eventNames.length; i++) {
        triggerEvent(elt, eventNames[i].trim(), []);
      }
    }
  }
  const WHITESPACE_OR_COMMA = /[\s,]/;
  const SYMBOL_START = /[_$a-zA-Z]/;
  const SYMBOL_CONT = /[_$a-zA-Z0-9]/;
  const STRINGISH_START = ['"', "'", '/'];
  const NOT_WHITESPACE = /[^\s]/;
  const COMBINED_SELECTOR_START = /[{(]/;
  const COMBINED_SELECTOR_END = /[})]/;

  /**
   * @param {string} str
   * @returns {string[]}
   */
  function tokenizeString(str) {
    /** @type string[] */
    const tokens = [];
    let position = 0;
    while (position < str.length) {
      if (SYMBOL_START.exec(str.charAt(position))) {
        var startPosition = position;
        while (SYMBOL_CONT.exec(str.charAt(position + 1))) {
          position++;
        }
        tokens.push(str.substring(startPosition, position + 1));
      } else if (STRINGISH_START.indexOf(str.charAt(position)) !== -1) {
        const startChar = str.charAt(position);
        var startPosition = position;
        position++;
        while (position < str.length && str.charAt(position) !== startChar) {
          if (str.charAt(position) === '\\') {
            position++;
          }
          position++;
        }
        tokens.push(str.substring(startPosition, position + 1));
      } else {
        const symbol = str.charAt(position);
        tokens.push(symbol);
      }
      position++;
    }
    return tokens
  }

  /**
   * @param {string} token
   * @param {string|null} last
   * @param {string} paramName
   * @returns {boolean}
   */
  function isPossibleRelativeReference(token, last, paramName) {
    return SYMBOL_START.exec(token.charAt(0)) &&
      token !== 'true' &&
      token !== 'false' &&
      token !== 'this' &&
      token !== paramName &&
      last !== '.'
  }

  /**
   * @param {EventTarget|string} elt
   * @param {string[]} tokens
   * @param {string} paramName
   * @returns {ConditionalFunction|null}
   */
  function maybeGenerateConditional(elt, tokens, paramName) {
    if (tokens[0] === '[') {
      tokens.shift();
      let bracketCount = 1;
      let conditionalSource = ' return (function(' + paramName + '){ return (';
      let last = null;
      while (tokens.length > 0) {
        const token = tokens[0];
        // @ts-ignore For some reason tsc doesn't understand the shift call, and thinks we're comparing the same value here, i.e. '[' vs ']'
        if (token === ']') {
          bracketCount--;
          if (bracketCount === 0) {
            if (last === null) {
              conditionalSource = conditionalSource + 'true';
            }
            tokens.shift();
            conditionalSource += ')})';
            try {
              const conditionFunction = maybeEval(elt, function() {
                return Function(conditionalSource)()
              },
              function() { return true });
              conditionFunction.source = conditionalSource;
              return conditionFunction
            } catch (e) {
              triggerErrorEvent(getDocument().body, 'htmx:syntax:error', { error: e, source: conditionalSource });
              return null
            }
          }
        } else if (token === '[') {
          bracketCount++;
        }
        if (isPossibleRelativeReference(token, last, paramName)) {
          conditionalSource += '((' + paramName + '.' + token + ') ? (' + paramName + '.' + token + ') : (window.' + token + '))';
        } else {
          conditionalSource = conditionalSource + token;
        }
        last = tokens.shift();
      }
    }
  }

  /**
   * @param {string[]} tokens
   * @param {RegExp} match
   * @returns {string}
   */
  function consumeUntil(tokens, match) {
    let result = '';
    while (tokens.length > 0 && !match.test(tokens[0])) {
      result += tokens.shift();
    }
    return result
  }

  /**
   * @param {string[]} tokens
   * @returns {string}
   */
  function consumeCSSSelector(tokens) {
    let result;
    if (tokens.length > 0 && COMBINED_SELECTOR_START.test(tokens[0])) {
      tokens.shift();
      result = consumeUntil(tokens, COMBINED_SELECTOR_END).trim();
      tokens.shift();
    } else {
      result = consumeUntil(tokens, WHITESPACE_OR_COMMA);
    }
    return result
  }

  const INPUT_SELECTOR = 'input, textarea, select';

  /**
   * @param {Element} elt
   * @param {string} explicitTrigger
   * @param {Object} cache for trigger specs
   * @returns {HtmxTriggerSpecification[]}
   */
  function parseAndCacheTrigger(elt, explicitTrigger, cache) {
    /** @type HtmxTriggerSpecification[] */
    const triggerSpecs = [];
    const tokens = tokenizeString(explicitTrigger);
    do {
      consumeUntil(tokens, NOT_WHITESPACE);
      const initialLength = tokens.length;
      const trigger = consumeUntil(tokens, /[,\[\s]/);
      if (trigger !== '') {
        if (trigger === 'every') {
          /** @type HtmxTriggerSpecification */
          const every = { trigger: 'every' };
          consumeUntil(tokens, NOT_WHITESPACE);
          every.pollInterval = parseInterval(consumeUntil(tokens, /[,\[\s]/));
          consumeUntil(tokens, NOT_WHITESPACE);
          var eventFilter = maybeGenerateConditional(elt, tokens, 'event');
          if (eventFilter) {
            every.eventFilter = eventFilter;
          }
          triggerSpecs.push(every);
        } else {
          /** @type HtmxTriggerSpecification */
          const triggerSpec = { trigger };
          var eventFilter = maybeGenerateConditional(elt, tokens, 'event');
          if (eventFilter) {
            triggerSpec.eventFilter = eventFilter;
          }
          consumeUntil(tokens, NOT_WHITESPACE);
          while (tokens.length > 0 && tokens[0] !== ',') {
            const token = tokens.shift();
            if (token === 'changed') {
              triggerSpec.changed = true;
            } else if (token === 'once') {
              triggerSpec.once = true;
            } else if (token === 'consume') {
              triggerSpec.consume = true;
            } else if (token === 'delay' && tokens[0] === ':') {
              tokens.shift();
              triggerSpec.delay = parseInterval(consumeUntil(tokens, WHITESPACE_OR_COMMA));
            } else if (token === 'from' && tokens[0] === ':') {
              tokens.shift();
              if (COMBINED_SELECTOR_START.test(tokens[0])) {
                var from_arg = consumeCSSSelector(tokens);
              } else {
                var from_arg = consumeUntil(tokens, WHITESPACE_OR_COMMA);
                if (from_arg === 'closest' || from_arg === 'find' || from_arg === 'next' || from_arg === 'previous') {
                  tokens.shift();
                  const selector = consumeCSSSelector(tokens);
                  // `next` and `previous` allow a selector-less syntax
                  if (selector.length > 0) {
                    from_arg += ' ' + selector;
                  }
                }
              }
              triggerSpec.from = from_arg;
            } else if (token === 'target' && tokens[0] === ':') {
              tokens.shift();
              triggerSpec.target = consumeCSSSelector(tokens);
            } else if (token === 'throttle' && tokens[0] === ':') {
              tokens.shift();
              triggerSpec.throttle = parseInterval(consumeUntil(tokens, WHITESPACE_OR_COMMA));
            } else if (token === 'queue' && tokens[0] === ':') {
              tokens.shift();
              triggerSpec.queue = consumeUntil(tokens, WHITESPACE_OR_COMMA);
            } else if (token === 'root' && tokens[0] === ':') {
              tokens.shift();
              triggerSpec[token] = consumeCSSSelector(tokens);
            } else if (token === 'threshold' && tokens[0] === ':') {
              tokens.shift();
              triggerSpec[token] = consumeUntil(tokens, WHITESPACE_OR_COMMA);
            } else {
              triggerErrorEvent(elt, 'htmx:syntax:error', { token: tokens.shift() });
            }
            consumeUntil(tokens, NOT_WHITESPACE);
          }
          triggerSpecs.push(triggerSpec);
        }
      }
      if (tokens.length === initialLength) {
        triggerErrorEvent(elt, 'htmx:syntax:error', { token: tokens.shift() });
      }
      consumeUntil(tokens, NOT_WHITESPACE);
    } while (tokens[0] === ',' && tokens.shift())
    if (cache) {
      cache[explicitTrigger] = triggerSpecs;
    }
    return triggerSpecs
  }

  /**
   * @param {Element} elt
   * @returns {HtmxTriggerSpecification[]}
   */
  function getTriggerSpecs(elt) {
    const explicitTrigger = getAttributeValue(elt, 'hx-trigger');
    let triggerSpecs = [];
    if (explicitTrigger) {
      const cache = htmx.config.triggerSpecsCache;
      triggerSpecs = (cache && cache[explicitTrigger]) || parseAndCacheTrigger(elt, explicitTrigger, cache);
    }

    if (triggerSpecs.length > 0) {
      return triggerSpecs
    } else if (matches(elt, 'form')) {
      return [{ trigger: 'submit' }]
    } else if (matches(elt, 'input[type="button"], input[type="submit"]')) {
      return [{ trigger: 'click' }]
    } else if (matches(elt, INPUT_SELECTOR)) {
      return [{ trigger: 'change' }]
    } else {
      return [{ trigger: 'click' }]
    }
  }

  /**
   * @param {Element} elt
   */
  function cancelPolling(elt) {
    getInternalData(elt).cancelled = true;
  }

  /**
   * @param {Element} elt
   * @param {TriggerHandler} handler
   * @param {HtmxTriggerSpecification} spec
   */
  function processPolling(elt, handler, spec) {
    const nodeData = getInternalData(elt);
    nodeData.timeout = getWindow().setTimeout(function() {
      if (bodyContains(elt) && nodeData.cancelled !== true) {
        if (!maybeFilterEvent(spec, elt, makeEvent('hx:poll:trigger', {
          triggerSpec: spec,
          target: elt
        }))) {
          handler(elt);
        }
        processPolling(elt, handler, spec);
      }
    }, spec.pollInterval);
  }

  /**
   * @param {HTMLAnchorElement} elt
   * @returns {boolean}
   */
  function isLocalLink(elt) {
    return location.hostname === elt.hostname &&
      getRawAttribute(elt, 'href') &&
      getRawAttribute(elt, 'href').indexOf('#') !== 0
  }

  /**
   * @param {Element} elt
   */
  function eltIsDisabled(elt) {
    return closest(elt, htmx.config.disableSelector)
  }

  /**
   * @param {Element} elt
   * @param {HtmxNodeInternalData} nodeData
   * @param {HtmxTriggerSpecification[]} triggerSpecs
   */
  function boostElement(elt, nodeData, triggerSpecs) {
    if ((elt instanceof HTMLAnchorElement && isLocalLink(elt) && (elt.target === '' || elt.target === '_self')) || (elt.tagName === 'FORM' && String(getRawAttribute(elt, 'method')).toLowerCase() !== 'dialog')) {
      nodeData.boosted = true;
      let verb, path;
      if (elt.tagName === 'A') {
        verb = (/** @type HttpVerb */('get'));
        path = getRawAttribute(elt, 'href');
      } else {
        const rawAttribute = getRawAttribute(elt, 'method');
        verb = (/** @type HttpVerb */(rawAttribute ? rawAttribute.toLowerCase() : 'get'));
        path = getRawAttribute(elt, 'action');
        if (path == null || path === '') {
          // if there is no action attribute on the form set path to current href before the
          // following logic to properly clear parameters on a GET (not on a POST!)
          path = location.href;
        }
        if (verb === 'get' && path.includes('?')) {
          path = path.replace(/\?[^#]+/, '');
        }
      }
      triggerSpecs.forEach(function(triggerSpec) {
        addEventListener(elt, function(node, evt) {
          const elt = asElement(node);
          if (eltIsDisabled(elt)) {
            cleanUpElement(elt);
            return
          }
          issueAjaxRequest(verb, path, elt, evt);
        }, nodeData, triggerSpec, true);
      });
    }
  }

  /**
   * @param {Event} evt
   * @param {Element} elt
   * @returns {boolean}
   */
  function shouldCancel(evt, elt) {
    if (evt.type === 'submit' && elt.tagName === 'FORM') {
      return true
    } else if (evt.type === 'click') {
      // find button wrapping the trigger element
      const btn = /** @type {HTMLButtonElement|HTMLInputElement|null} */ (elt.closest('input[type="submit"], button'));
      // Do not cancel on buttons that 1) don't have a related form or 2) have a type attribute of 'reset'/'button'.
      if (btn && btn.form && btn.type === 'submit') {
        return true
      }

      // find link wrapping the trigger element
      const link = elt.closest('a');
      // Allow links with href="#fragment" (anchors with content after #) to perform normal fragment navigation.
      // Cancel default action for links with href="#" (bare hash) to prevent scrolling to top and unwanted URL changes.
      const samePageAnchor = /^#.+/;
      if (link && link.href && !samePageAnchor.test(link.getAttribute('href'))) {
        return true
      }
    }
    return false
  }

  /**
   * @param {Node} elt
   * @param {Event|MouseEvent|KeyboardEvent|TouchEvent} evt
   * @returns {boolean}
   */
  function ignoreBoostedAnchorCtrlClick(elt, evt) {
    return getInternalData(elt).boosted && elt instanceof HTMLAnchorElement && evt.type === 'click' &&
      // @ts-ignore this will resolve to undefined for events that don't define those properties, which is fine
      (evt.ctrlKey || evt.metaKey)
  }

  /**
   * @param {HtmxTriggerSpecification} triggerSpec
   * @param {Node} elt
   * @param {Event} evt
   * @returns {boolean}
   */
  function maybeFilterEvent(triggerSpec, elt, evt) {
    const eventFilter = triggerSpec.eventFilter;
    if (eventFilter) {
      try {
        return eventFilter.call(elt, evt) !== true
      } catch (e) {
        const source = eventFilter.source;
        triggerErrorEvent(getDocument().body, 'htmx:eventFilter:error', { error: e, source });
        return true
      }
    }
    return false
  }

  /**
   * @param {Element} elt
   * @param {TriggerHandler} handler
   * @param {HtmxNodeInternalData} nodeData
   * @param {HtmxTriggerSpecification} triggerSpec
   * @param {boolean} [explicitCancel]
   */
  function addEventListener(elt, handler, nodeData, triggerSpec, explicitCancel) {
    const elementData = getInternalData(elt);
    /** @type {(Node|Window)[]} */
    let eltsToListenOn;
    if (triggerSpec.from) {
      eltsToListenOn = querySelectorAllExt(elt, triggerSpec.from);
    } else {
      eltsToListenOn = [elt];
    }
    // store the initial values of the elements, so we can tell if they change
    if (triggerSpec.changed) {
      if (!('lastValue' in elementData)) {
        elementData.lastValue = new WeakMap();
      }
      eltsToListenOn.forEach(function(eltToListenOn) {
        if (!elementData.lastValue.has(triggerSpec)) {
          elementData.lastValue.set(triggerSpec, new WeakMap());
        }
        // @ts-ignore value will be undefined for non-input elements, which is fine
        elementData.lastValue.get(triggerSpec).set(eltToListenOn, eltToListenOn.value);
      });
    }
    forEach(eltsToListenOn, function(eltToListenOn) {
      /** @type EventListener */
      const eventListener = function(evt) {
        if (!bodyContains(elt)) {
          eltToListenOn.removeEventListener(triggerSpec.trigger, eventListener);
          return
        }
        if (ignoreBoostedAnchorCtrlClick(elt, evt)) {
          return
        }
        if (explicitCancel || shouldCancel(evt, eltToListenOn)) {
          evt.preventDefault();
        }
        if (maybeFilterEvent(triggerSpec, elt, evt)) {
          return
        }
        const eventData = getInternalData(evt);
        eventData.triggerSpec = triggerSpec;
        if (eventData.handledFor == null) {
          eventData.handledFor = [];
        }
        if (eventData.handledFor.indexOf(elt) < 0) {
          eventData.handledFor.push(elt);
          if (triggerSpec.consume) {
            evt.stopPropagation();
          }
          if (triggerSpec.target && evt.target) {
            if (!matches(asElement(evt.target), triggerSpec.target)) {
              return
            }
          }
          if (triggerSpec.once) {
            if (elementData.triggeredOnce) {
              return
            } else {
              elementData.triggeredOnce = true;
            }
          }
          if (triggerSpec.changed) {
            const node = evt.target;
            // @ts-ignore value will be undefined for non-input elements, which is fine
            const value = node.value;
            const lastValue = elementData.lastValue.get(triggerSpec);
            if (lastValue.has(node) && lastValue.get(node) === value) {
              return
            }
            lastValue.set(node, value);
          }
          if (elementData.delayed) {
            clearTimeout(elementData.delayed);
          }
          if (elementData.throttle) {
            return
          }

          if (triggerSpec.throttle > 0) {
            if (!elementData.throttle) {
              triggerEvent(elt, 'htmx:trigger');
              handler(elt, evt);
              elementData.throttle = getWindow().setTimeout(function() {
                elementData.throttle = null;
              }, triggerSpec.throttle);
            }
          } else if (triggerSpec.delay > 0) {
            elementData.delayed = getWindow().setTimeout(function() {
              triggerEvent(elt, 'htmx:trigger');
              handler(elt, evt);
            }, triggerSpec.delay);
          } else {
            triggerEvent(elt, 'htmx:trigger');
            handler(elt, evt);
          }
        }
      };
      if (nodeData.listenerInfos == null) {
        nodeData.listenerInfos = [];
      }
      nodeData.listenerInfos.push({
        trigger: triggerSpec.trigger,
        listener: eventListener,
        on: eltToListenOn
      });
      eltToListenOn.addEventListener(triggerSpec.trigger, eventListener);
    });
  }

  let windowIsScrolling = false; // used by initScrollHandler
  let scrollHandler = null;
  function initScrollHandler() {
    if (!scrollHandler) {
      scrollHandler = function() {
        windowIsScrolling = true;
      };
      window.addEventListener('scroll', scrollHandler);
      window.addEventListener('resize', scrollHandler);
      setInterval(function() {
        if (windowIsScrolling) {
          windowIsScrolling = false;
          forEach(getDocument().querySelectorAll("[hx-trigger*='revealed'],[data-hx-trigger*='revealed']"), function(elt) {
            maybeReveal(elt);
          });
        }
      }, 200);
    }
  }

  /**
   * @param {Element} elt
   */
  function maybeReveal(elt) {
    if (!hasAttribute(elt, 'data-hx-revealed') && isScrolledIntoView(elt)) {
      elt.setAttribute('data-hx-revealed', 'true');
      const nodeData = getInternalData(elt);
      if (nodeData.initHash) {
        triggerEvent(elt, 'revealed');
      } else {
        // if the node isn't initialized, wait for it before triggering the request
        elt.addEventListener('htmx:afterProcessNode', function() { triggerEvent(elt, 'revealed'); }, { once: true });
      }
    }
  }

  //= ===================================================================

  /**
   * @param {Element} elt
   * @param {TriggerHandler} handler
   * @param {HtmxNodeInternalData} nodeData
   * @param {number} delay
   */
  function loadImmediately(elt, handler, nodeData, delay) {
    const load = function() {
      if (!nodeData.loaded) {
        nodeData.loaded = true;
        triggerEvent(elt, 'htmx:trigger');
        handler(elt);
      }
    };
    if (delay > 0) {
      getWindow().setTimeout(load, delay);
    } else {
      load();
    }
  }

  /**
   * @param {Element} elt
   * @param {HtmxNodeInternalData} nodeData
   * @param {HtmxTriggerSpecification[]} triggerSpecs
   * @returns {boolean}
   */
  function processVerbs(elt, nodeData, triggerSpecs) {
    let explicitAction = false;
    forEach(VERBS, function(verb) {
      if (hasAttribute(elt, 'hx-' + verb)) {
        const path = getAttributeValue(elt, 'hx-' + verb);
        explicitAction = true;
        nodeData.path = path;
        nodeData.verb = verb;
        triggerSpecs.forEach(function(triggerSpec) {
          addTriggerHandler(elt, triggerSpec, nodeData, function(node, evt) {
            const elt = asElement(node);
            if (eltIsDisabled(elt)) {
              cleanUpElement(elt);
              return
            }
            issueAjaxRequest(verb, path, elt, evt);
          });
        });
      }
    });
    return explicitAction
  }

  /**
   * @callback TriggerHandler
   * @param {Element} elt
   * @param {Event} [evt]
   */

  /**
   * @param {Element} elt
   * @param {HtmxTriggerSpecification} triggerSpec
   * @param {HtmxNodeInternalData} nodeData
   * @param {TriggerHandler} handler
   */
  function addTriggerHandler(elt, triggerSpec, nodeData, handler) {
    if (triggerSpec.trigger === 'revealed') {
      initScrollHandler();
      addEventListener(elt, handler, nodeData, triggerSpec);
      maybeReveal(asElement(elt));
    } else if (triggerSpec.trigger === 'intersect') {
      const observerOptions = {};
      if (triggerSpec.root) {
        observerOptions.root = querySelectorExt(elt, triggerSpec.root);
      }
      if (triggerSpec.threshold) {
        observerOptions.threshold = parseFloat(triggerSpec.threshold);
      }
      const observer = new IntersectionObserver(function(entries) {
        for (let i = 0; i < entries.length; i++) {
          const entry = entries[i];
          if (entry.isIntersecting) {
            triggerEvent(elt, 'intersect');
            break
          }
        }
      }, observerOptions);
      observer.observe(asElement(elt));
      addEventListener(asElement(elt), handler, nodeData, triggerSpec);
    } else if (!nodeData.firstInitCompleted && triggerSpec.trigger === 'load') {
      if (!maybeFilterEvent(triggerSpec, elt, makeEvent('load', { elt }))) {
        loadImmediately(asElement(elt), handler, nodeData, triggerSpec.delay);
      }
    } else if (triggerSpec.pollInterval > 0) {
      nodeData.polling = true;
      processPolling(asElement(elt), handler, triggerSpec);
    } else {
      addEventListener(elt, handler, nodeData, triggerSpec);
    }
  }

  /**
   * @param {Node} node
   * @returns {boolean}
   */
  function shouldProcessHxOn(node) {
    const elt = asElement(node);
    if (!elt) {
      return false
    }
    const attributes = elt.attributes;
    for (let j = 0; j < attributes.length; j++) {
      const attrName = attributes[j].name;
      if (startsWith(attrName, 'hx-on:') || startsWith(attrName, 'data-hx-on:') ||
        startsWith(attrName, 'hx-on-') || startsWith(attrName, 'data-hx-on-')) {
        return true
      }
    }
    return false
  }

  /**
   * @param {Node} elt
   * @returns {Element[]}
   */
  const HX_ON_QUERY = new XPathEvaluator()
    .createExpression('.//*[@*[ starts-with(name(), "hx-on:") or starts-with(name(), "data-hx-on:") or' +
      ' starts-with(name(), "hx-on-") or starts-with(name(), "data-hx-on-") ]]');

  function processHXOnRoot(elt, elements) {
    if (shouldProcessHxOn(elt)) {
      elements.push(asElement(elt));
    }
    const iter = HX_ON_QUERY.evaluate(elt);
    let node = null;
    while (node = iter.iterateNext()) elements.push(asElement(node));
  }

  function findHxOnWildcardElements(elt) {
    /** @type {Element[]} */
    const elements = [];
    if (elt instanceof DocumentFragment) {
      for (const child of elt.childNodes) {
        processHXOnRoot(child, elements);
      }
    } else {
      processHXOnRoot(elt, elements);
    }
    return elements
  }

  /**
   * @param {Element} elt
   * @returns {NodeListOf<Element>|[]}
   */
  function findElementsToProcess(elt) {
    if (elt.querySelectorAll) {
      const boostedSelector = ', [hx-boost] a, [data-hx-boost] a, a[hx-boost], a[data-hx-boost]';

      const extensionSelectors = [];
      for (const e in extensions) {
        const extension = extensions[e];
        if (extension.getSelectors) {
          var selectors = extension.getSelectors();
          if (selectors) {
            extensionSelectors.push(selectors);
          }
        }
      }

      const results = elt.querySelectorAll(VERB_SELECTOR + boostedSelector + ", form, [type='submit']," +
        ' [hx-ext], [data-hx-ext], [hx-trigger], [data-hx-trigger]' + extensionSelectors.flat().map(s => ', ' + s).join(''));

      return results
    } else {
      return []
    }
  }

  /**
   * Handle submit buttons/inputs that have the form attribute set
   * see https://developer.mozilla.org/docs/Web/HTML/Element/button
   * @param {Event} evt
   */
  function maybeSetLastButtonClicked(evt) {
    const elt = getTargetButton(evt.target);
    const internalData = getRelatedFormData(evt);
    if (internalData) {
      internalData.lastButtonClicked = elt;
    }
  }

  /**
   * @param {Event} evt
   */
  function maybeUnsetLastButtonClicked(evt) {
    const internalData = getRelatedFormData(evt);
    if (internalData) {
      internalData.lastButtonClicked = null;
    }
  }

  /**
   * @param {EventTarget} target
   * @returns {HTMLButtonElement|HTMLInputElement|null}
   */
  function getTargetButton(target) {
    return /** @type {HTMLButtonElement|HTMLInputElement|null} */ (closest(asElement(target), "button, input[type='submit']"))
  }

  /**
   * @param {Element} elt
   * @returns {HTMLFormElement|null}
   */
  function getRelatedForm(elt) {
    // @ts-ignore Get the related form if available, else find the closest parent form
    return elt.form || closest(elt, 'form')
  }

  /**
   * @param {Event} evt
   * @returns {HtmxNodeInternalData|undefined}
   */
  function getRelatedFormData(evt) {
    const elt = getTargetButton(evt.target);
    if (!elt) {
      return
    }
    const form = getRelatedForm(elt);
    if (!form) {
      return
    }
    return getInternalData(form)
  }

  /**
   * @param {EventTarget} elt
   */
  function initButtonTracking(elt) {
    // need to handle both click and focus in:
    //   focusin - in case someone tabs in to a button and hits the space bar
    //   click - on OSX buttons do not focus on click see https://bugs.webkit.org/show_bug.cgi?id=13724
    elt.addEventListener('click', maybeSetLastButtonClicked);
    elt.addEventListener('focusin', maybeSetLastButtonClicked);
    elt.addEventListener('focusout', maybeUnsetLastButtonClicked);
  }

  /**
   * @param {Element} elt
   * @param {string} eventName
   * @param {string} code
   */
  function addHxOnEventHandler(elt, eventName, code) {
    const nodeData = getInternalData(elt);
    if (!Array.isArray(nodeData.onHandlers)) {
      nodeData.onHandlers = [];
    }
    let func;
    /** @type EventListener */
    const listener = function(e) {
      maybeEval(elt, function() {
        if (eltIsDisabled(elt)) {
          return
        }
        if (!func) {
          func = new Function('event', code);
        }
        func.call(elt, e);
      });
    };
    elt.addEventListener(eventName, listener);
    nodeData.onHandlers.push({ event: eventName, listener });
  }

  /**
   * @param {Element} elt
   */
  function processHxOnWildcard(elt) {
    // wipe any previous on handlers so that this function takes precedence
    deInitOnHandlers(elt);

    for (let i = 0; i < elt.attributes.length; i++) {
      const name = elt.attributes[i].name;
      const value = elt.attributes[i].value;
      if (startsWith(name, 'hx-on') || startsWith(name, 'data-hx-on')) {
        const afterOnPosition = name.indexOf('-on') + 3;
        const nextChar = name.slice(afterOnPosition, afterOnPosition + 1);
        if (nextChar === '-' || nextChar === ':') {
          let eventName = name.slice(afterOnPosition + 1);
          // if the eventName starts with a colon or dash, prepend "htmx" for shorthand support
          if (startsWith(eventName, ':')) {
            eventName = 'htmx' + eventName;
          } else if (startsWith(eventName, '-')) {
            eventName = 'htmx:' + eventName.slice(1);
          } else if (startsWith(eventName, 'htmx-')) {
            eventName = 'htmx:' + eventName.slice(5);
          }

          addHxOnEventHandler(elt, eventName, value);
        }
      }
    }
  }

  /**
   * @param {Element|HTMLInputElement} elt
   */
  function initNode(elt) {
    triggerEvent(elt, 'htmx:beforeProcessNode');

    const nodeData = getInternalData(elt);
    const triggerSpecs = getTriggerSpecs(elt);
    const hasExplicitHttpAction = processVerbs(elt, nodeData, triggerSpecs);

    if (!hasExplicitHttpAction) {
      if (getClosestAttributeValue(elt, 'hx-boost') === 'true') {
        boostElement(elt, nodeData, triggerSpecs);
      } else if (hasAttribute(elt, 'hx-trigger')) {
        triggerSpecs.forEach(function(triggerSpec) {
          // For "naked" triggers, don't do anything at all
          addTriggerHandler(elt, triggerSpec, nodeData, function() {
          });
        });
      }
    }

    // Handle submit buttons/inputs that have the form attribute set
    // see https://developer.mozilla.org/docs/Web/HTML/Element/button
    if (elt.tagName === 'FORM' || (getRawAttribute(elt, 'type') === 'submit' && hasAttribute(elt, 'form'))) {
      initButtonTracking(elt);
    }

    nodeData.firstInitCompleted = true;
    triggerEvent(elt, 'htmx:afterProcessNode');
  }

  /**
   * @param {Element} elt
   * @returns {boolean}
   */
  function maybeDeInitAndHash(elt) {
    // Ensure only valid Elements and not shadow DOM roots are inited
    if (!(elt instanceof Element)) {
      return false
    }

    const nodeData = getInternalData(elt);
    const hash = attributeHash(elt);
    if (nodeData.initHash !== hash) {
      deInitNode(elt);
      nodeData.initHash = hash;
      return true
    }
    return false
  }

  /**
   * Processes new content, enabling htmx behavior. This can be useful if you have content that is added to the DOM outside of the normal htmx request cycle but still want htmx attributes to work.
   *
   * @see https://htmx.org/api/#process
   *
   * @param {Element|string} elt element to process
   */
  function processNode(elt) {
    elt = resolveTarget(elt);
    if (eltIsDisabled(elt)) {
      cleanUpElement(elt);
      return
    }

    const elementsToInit = [];
    if (maybeDeInitAndHash(elt)) {
      elementsToInit.push(elt);
    }
    forEach(findElementsToProcess(elt), function(child) {
      if (eltIsDisabled(child)) {
        cleanUpElement(child);
        return
      }
      if (maybeDeInitAndHash(child)) {
        elementsToInit.push(child);
      }
    });

    forEach(findHxOnWildcardElements(elt), processHxOnWildcard);
    forEach(elementsToInit, initNode);
  }

  //= ===================================================================
  // Event/Log Support
  //= ===================================================================

  /**
   * @param {string} str
   * @returns {string}
   */
  function kebabEventName(str) {
    return str.replace(/([a-z0-9])([A-Z])/g, '$1-$2').toLowerCase()
  }

  /**
   * @param {string} eventName
   * @param {any} detail
   * @returns {CustomEvent}
   */
  function makeEvent(eventName, detail) {
    // TODO: `composed: true` here is a hack to make global event handlers work with events in shadow DOM
    // This breaks expected encapsulation but needs to be here until decided otherwise by core devs
    return new CustomEvent(eventName, { bubbles: true, cancelable: true, composed: true, detail })
  }

  /**
   * @param {EventTarget|string} elt
   * @param {string} eventName
   * @param {any=} detail
   */
  function triggerErrorEvent(elt, eventName, detail) {
    triggerEvent(elt, eventName, mergeObjects({ error: eventName }, detail));
  }

  /**
   * @param {string} eventName
   * @returns {boolean}
   */
  function ignoreEventForLogging(eventName) {
    return eventName === 'htmx:afterProcessNode'
  }

  /**
   * `withExtensions` locates all active extensions for a provided element, then
   * executes the provided function using each of the active extensions. You can filter
   * the element's extensions by giving it a list of extensions to ignore. It should
   * be called internally at every extendable execution point in htmx.
   *
   * @param {Element} elt
   * @param {(extension:HtmxExtension) => void} toDo
   * @param {string[]=} extensionsToIgnore
   * @returns void
   */
  function withExtensions(elt, toDo, extensionsToIgnore) {
    forEach(getExtensions(elt, [], extensionsToIgnore), function(extension) {
      try {
        toDo(extension);
      } catch (e) {
        logError(e);
      }
    });
  }

  function logError(msg) {
    console.error(msg);
  }

  /**
   * Triggers a given event on an element
   *
   * @see https://htmx.org/api/#trigger
   *
   * @param {EventTarget|string} elt the element to trigger the event on
   * @param {string} eventName the name of the event to trigger
   * @param {any=} detail details for the event
   * @returns {boolean}
   */
  function triggerEvent(elt, eventName, detail) {
    elt = resolveTarget(elt);
    if (detail == null) {
      detail = {};
    }
    detail.elt = elt;
    const event = makeEvent(eventName, detail);
    if (htmx.logger && !ignoreEventForLogging(eventName)) {
      htmx.logger(elt, eventName, detail);
    }
    if (detail.error) {
      logError(detail.error);
      triggerEvent(elt, 'htmx:error', { errorInfo: detail });
    }
    let eventResult = elt.dispatchEvent(event);
    const kebabName = kebabEventName(eventName);
    if (eventResult && kebabName !== eventName) {
      const kebabedEvent = makeEvent(kebabName, event.detail);
      eventResult = eventResult && elt.dispatchEvent(kebabedEvent);
    }
    withExtensions(asElement(elt), function(extension) {
      eventResult = eventResult && (extension.onEvent(eventName, event) !== false && !event.defaultPrevented);
    });
    return eventResult
  }

  //= ===================================================================
  // History Support
  //= ===================================================================
  let currentPathForHistory;

  /**
   * @param {string} path
   */
  function setCurrentPathForHistory(path) {
    currentPathForHistory = path;
    if (canAccessLocalStorage()) {
      sessionStorage.setItem('htmx-current-path-for-history', path);
    }
  }

  setCurrentPathForHistory(location.pathname + location.search);

  /**
   * @returns {Element}
   */
  function getHistoryElement() {
    const historyElt = getDocument().querySelector('[hx-history-elt],[data-hx-history-elt]');
    return historyElt || getDocument().body
  }

  /**
   * @param {string} url
   * @param {Element} rootElt
   */
  function saveToHistoryCache(url, rootElt) {
    if (!canAccessLocalStorage()) {
      return
    }

    // get state to save
    const innerHTML = cleanInnerHtmlForHistory(rootElt);
    const title = getDocument().title;
    const scroll = window.scrollY;

    if (htmx.config.historyCacheSize <= 0) {
      // make sure that an eventually already existing cache is purged
      sessionStorage.removeItem('htmx-history-cache');
      return
    }

    url = normalizePath(url);

    const historyCache = parseJSON(sessionStorage.getItem('htmx-history-cache')) || [];
    for (let i = 0; i < historyCache.length; i++) {
      if (historyCache[i].url === url) {
        historyCache.splice(i, 1);
        break
      }
    }

    /** @type HtmxHistoryItem */
    const newHistoryItem = { url, content: innerHTML, title, scroll };

    triggerEvent(getDocument().body, 'htmx:historyItemCreated', { item: newHistoryItem, cache: historyCache });

    historyCache.push(newHistoryItem);
    while (historyCache.length > htmx.config.historyCacheSize) {
      historyCache.shift();
    }

    // keep trying to save the cache until it succeeds or is empty
    while (historyCache.length > 0) {
      try {
        sessionStorage.setItem('htmx-history-cache', JSON.stringify(historyCache));
        break
      } catch (e) {
        triggerErrorEvent(getDocument().body, 'htmx:historyCacheError', { cause: e, cache: historyCache });
        historyCache.shift(); // shrink the cache and retry
      }
    }
  }

  /**
   * @typedef {Object} HtmxHistoryItem
   * @property {string} url
   * @property {string} content
   * @property {string} title
   * @property {number} scroll
   */

  /**
   * @param {string} url
   * @returns {HtmxHistoryItem|null}
   */
  function getCachedHistory(url) {
    if (!canAccessLocalStorage()) {
      return null
    }

    url = normalizePath(url);

    const historyCache = parseJSON(sessionStorage.getItem('htmx-history-cache')) || [];
    for (let i = 0; i < historyCache.length; i++) {
      if (historyCache[i].url === url) {
        return historyCache[i]
      }
    }
    return null
  }

  /**
   * @param {Element} elt
   * @returns {string}
   */
  function cleanInnerHtmlForHistory(elt) {
    const className = htmx.config.requestClass;
    const clone = /** @type Element */ (elt.cloneNode(true));
    forEach(findAll(clone, '.' + className), function(child) {
      removeClassFromElement(child, className);
    });
    // remove the disabled attribute for any element disabled due to an htmx request
    forEach(findAll(clone, '[data-disabled-by-htmx]'), function(child) {
      child.removeAttribute('disabled');
    });
    return clone.innerHTML
  }

  function saveCurrentPageToHistory() {
    const elt = getHistoryElement();
    let path = currentPathForHistory;
    if (canAccessLocalStorage()) {
      path = sessionStorage.getItem('htmx-current-path-for-history');
    }
    path = path || location.pathname + location.search;

    // Allow history snapshot feature to be disabled where hx-history="false"
    // is present *anywhere* in the current document we're about to save,
    // so we can prevent privileged data entering the cache.
    // The page will still be reachable as a history entry, but htmx will fetch it
    // live from the server onpopstate rather than look in the sessionStorage cache
    const disableHistoryCache = getDocument().querySelector('[hx-history="false" i],[data-hx-history="false" i]');
    if (!disableHistoryCache) {
      triggerEvent(getDocument().body, 'htmx:beforeHistorySave', { path, historyElt: elt });
      saveToHistoryCache(path, elt);
    }

    if (htmx.config.historyEnabled) history.replaceState({ htmx: true }, getDocument().title, location.href);
  }

  /**
   * @param {string} path
   */
  function pushUrlIntoHistory(path) {
  // remove the cache buster parameter, if any
    if (htmx.config.getCacheBusterParam) {
      path = path.replace(/org\.htmx\.cache-buster=[^&]*&?/, '');
      if (endsWith(path, '&') || endsWith(path, '?')) {
        path = path.slice(0, -1);
      }
    }
    if (htmx.config.historyEnabled) {
      history.pushState({ htmx: true }, '', path);
    }
    setCurrentPathForHistory(path);
  }

  /**
   * @param {string} path
   */
  function replaceUrlInHistory(path) {
    if (htmx.config.historyEnabled) history.replaceState({ htmx: true }, '', path);
    setCurrentPathForHistory(path);
  }

  /**
   * @param {HtmxSettleTask[]} tasks
   */
  function settleImmediately(tasks) {
    forEach(tasks, function(task) {
      task.call(undefined);
    });
  }

  /**
   * @param {string} path
   */
  function loadHistoryFromServer(path) {
    const request = new XMLHttpRequest();
    const swapSpec = { swapStyle: 'innerHTML', swapDelay: 0, settleDelay: 0 };
    const details = { path, xhr: request, historyElt: getHistoryElement(), swapSpec };
    request.open('GET', path, true);
    if (htmx.config.historyRestoreAsHxRequest) {
      request.setRequestHeader('HX-Request', 'true');
    }
    request.setRequestHeader('HX-History-Restore-Request', 'true');
    request.setRequestHeader('HX-Current-URL', location.href);
    request.onload = function() {
      if (this.status >= 200 && this.status < 400) {
        details.response = this.response;
        triggerEvent(getDocument().body, 'htmx:historyCacheMissLoad', details);
        swap(details.historyElt, details.response, swapSpec, {
          contextElement: details.historyElt,
          historyRequest: true
        });
        setCurrentPathForHistory(details.path);
        triggerEvent(getDocument().body, 'htmx:historyRestore', { path, cacheMiss: true, serverResponse: details.response });
      } else {
        triggerErrorEvent(getDocument().body, 'htmx:historyCacheMissLoadError', details);
      }
    };
    if (triggerEvent(getDocument().body, 'htmx:historyCacheMiss', details)) {
      request.send(); // only send request if event not prevented
    }
  }

  /**
   * @param {string} [path]
   */
  function restoreHistory(path) {
    saveCurrentPageToHistory();
    path = path || location.pathname + location.search;
    const cached = getCachedHistory(path);
    if (cached) {
      const swapSpec = { swapStyle: 'innerHTML', swapDelay: 0, settleDelay: 0, scroll: cached.scroll };
      const details = { path, item: cached, historyElt: getHistoryElement(), swapSpec };
      if (triggerEvent(getDocument().body, 'htmx:historyCacheHit', details)) {
        swap(details.historyElt, cached.content, swapSpec, {
          contextElement: details.historyElt,
          title: cached.title
        });
        setCurrentPathForHistory(details.path);
        triggerEvent(getDocument().body, 'htmx:historyRestore', details);
      }
    } else {
      if (htmx.config.refreshOnHistoryMiss) {
        // @ts-ignore: optional parameter in reload() function throws error
        // noinspection JSUnresolvedReference
        htmx.location.reload(true);
      } else {
        loadHistoryFromServer(path);
      }
    }
  }

  /**
   * @param {Element} elt
   * @returns {Element[]}
   */
  function addRequestIndicatorClasses(elt) {
    let indicators = /** @type Element[] */ (findAttributeTargets(elt, 'hx-indicator'));
    if (indicators == null) {
      indicators = [elt];
    }
    forEach(indicators, function(ic) {
      const internalData = getInternalData(ic);
      internalData.requestCount = (internalData.requestCount || 0) + 1;
      ic.classList.add.call(ic.classList, htmx.config.requestClass);
    });
    return indicators
  }

  /**
   * @param {Element} elt
   * @returns {Element[]}
   */
  function disableElements(elt) {
    let disabledElts = /** @type Element[] */ (findAttributeTargets(elt, 'hx-disabled-elt'));
    if (disabledElts == null) {
      disabledElts = [];
    }
    forEach(disabledElts, function(disabledElement) {
      const internalData = getInternalData(disabledElement);
      internalData.requestCount = (internalData.requestCount || 0) + 1;
      disabledElement.setAttribute('disabled', '');
      disabledElement.setAttribute('data-disabled-by-htmx', '');
    });
    return disabledElts
  }

  /**
   * @param {Element[]} indicators
   * @param {Element[]} disabled
   */
  function removeRequestIndicators(indicators, disabled) {
    forEach(indicators.concat(disabled), function(ele) {
      const internalData = getInternalData(ele);
      internalData.requestCount = (internalData.requestCount || 1) - 1;
    });
    forEach(indicators, function(ic) {
      const internalData = getInternalData(ic);
      if (internalData.requestCount === 0) {
        ic.classList.remove.call(ic.classList, htmx.config.requestClass);
      }
    });
    forEach(disabled, function(disabledElement) {
      const internalData = getInternalData(disabledElement);
      if (internalData.requestCount === 0) {
        disabledElement.removeAttribute('disabled');
        disabledElement.removeAttribute('data-disabled-by-htmx');
      }
    });
  }

  //= ===================================================================
  // Input Value Processing
  //= ===================================================================

  /**
   * @param {Element[]} processed
   * @param {Element} elt
   * @returns {boolean}
   */
  function haveSeenNode(processed, elt) {
    for (let i = 0; i < processed.length; i++) {
      const node = processed[i];
      if (node.isSameNode(elt)) {
        return true
      }
    }
    return false
  }

  /**
   * @param {Element} element
   * @return {boolean}
   */
  function shouldInclude(element) {
    // Cast to trick tsc, undefined values will work fine here
    const elt = /** @type {HTMLInputElement} */ (element);
    if (elt.name === '' || elt.name == null || elt.disabled || closest(elt, 'fieldset[disabled]')) {
      return false
    }
    // ignore "submitter" types (see jQuery src/serialize.js)
    if (elt.type === 'button' || elt.type === 'submit' || elt.tagName === 'image' || elt.tagName === 'reset' || elt.tagName === 'file') {
      return false
    }
    if (elt.type === 'checkbox' || elt.type === 'radio') {
      return elt.checked
    }
    return true
  }

  /**
   * @param {string} name
   * @param {string|Array|FormDataEntryValue} value
   * @param {FormData} formData */
  function addValueToFormData(name, value, formData) {
    if (name != null && value != null) {
      if (Array.isArray(value)) {
        value.forEach(function(v) { formData.append(name, v); });
      } else {
        formData.append(name, value);
      }
    }
  }

  /**
   * @param {string} name
   * @param {string|Array} value
   * @param {FormData} formData */
  function removeValueFromFormData(name, value, formData) {
    if (name != null && value != null) {
      let values = formData.getAll(name);
      if (Array.isArray(value)) {
        values = values.filter(v => value.indexOf(v) < 0);
      } else {
        values = values.filter(v => v !== value);
      }
      formData.delete(name);
      forEach(values, v => formData.append(name, v));
    }
  }

  /**
   * @param {Element} elt
   * @returns {string|Array}
   */
  function getValueFromInput(elt) {
    if (elt instanceof HTMLSelectElement && elt.multiple) {
      return toArray(elt.querySelectorAll('option:checked')).map(function(e) { return (/** @type HTMLOptionElement */(e)).value })
    }
    // include file inputs
    if (elt instanceof HTMLInputElement && elt.files) {
      return toArray(elt.files)
    }
    // @ts-ignore value will be undefined for non-input elements, which is fine
    return elt.value
  }

  /**
   * @param {Element[]} processed
   * @param {FormData} formData
   * @param {HtmxElementValidationError[]} errors
   * @param {Element|HTMLInputElement|HTMLSelectElement|HTMLFormElement} elt
   * @param {boolean} validate
   */
  function processInputValue(processed, formData, errors, elt, validate) {
    if (elt == null || haveSeenNode(processed, elt)) {
      return
    } else {
      processed.push(elt);
    }
    if (shouldInclude(elt)) {
      const name = getRawAttribute(elt, 'name');
      addValueToFormData(name, getValueFromInput(elt), formData);
      if (validate) {
        validateElement(elt, errors);
      }
    }
    if (elt instanceof HTMLFormElement) {
      forEach(elt.elements, function(input) {
        if (processed.indexOf(input) >= 0) {
          // The input has already been processed and added to the values, but the FormData that will be
          //  constructed right after on the form, will include it once again. So remove that input's value
          //  now to avoid duplicates
          removeValueFromFormData(input.name, getValueFromInput(input), formData);
        } else {
          processed.push(input);
        }
        if (validate) {
          validateElement(input, errors);
        }
      });
      new FormData(elt).forEach(function(value, name) {
        if (value instanceof File && value.name === '') {
          return // ignore no-name files
        }
        addValueToFormData(name, value, formData);
      });
    }
  }

  /**
   * @param {Element} elt
   * @param {HtmxElementValidationError[]} errors
   */
  function validateElement(elt, errors) {
    const element = /** @type {HTMLElement & ElementInternals} */ (elt);
    if (element.willValidate) {
      triggerEvent(element, 'htmx:validation:validate');
      if (!element.checkValidity()) {
        if (
          triggerEvent(element, 'htmx:validation:failed', {
            message: element.validationMessage,
            validity: element.validity
          }) &&
          !errors.length &&
          htmx.config.reportValidityOfForms
        ) {
          element.reportValidity();
        }
        errors.push({ elt: element, message: element.validationMessage, validity: element.validity });
      }
    }
  }

  /**
   * Override values in the one FormData with those from another.
   * @param {FormData} receiver the formdata that will be mutated
   * @param {FormData} donor the formdata that will provide the overriding values
   * @returns {FormData} the {@linkcode receiver}
   */
  function overrideFormData(receiver, donor) {
    for (const key of donor.keys()) {
      receiver.delete(key);
    }
    donor.forEach(function(value, key) {
      receiver.append(key, value);
    });
    return receiver
  }

  /**
 * @param {Element|HTMLFormElement} elt
 * @param {HttpVerb} verb
 * @returns {{errors: HtmxElementValidationError[], formData: FormData, values: Object}}
 */
  function getInputValues(elt, verb) {
    /** @type Element[] */
    const processed = [];
    const formData = new FormData();
    const priorityFormData = new FormData();
    /** @type HtmxElementValidationError[] */
    const errors = [];
    const internalData = getInternalData(elt);
    if (internalData.lastButtonClicked && !bodyContains(internalData.lastButtonClicked)) {
      internalData.lastButtonClicked = null;
    }

    // only validate when form is directly submitted and novalidate or formnovalidate are not set
    // or if the element has an explicit hx-validate="true" on it
    let validate = (elt instanceof HTMLFormElement && elt.noValidate !== true) || getAttributeValue(elt, 'hx-validate') === 'true';
    if (internalData.lastButtonClicked) {
      validate = validate && internalData.lastButtonClicked.formNoValidate !== true;
    }

    // for a non-GET include the related form, which may or may not be a parent element of elt
    if (verb !== 'get') {
      processInputValue(processed, priorityFormData, errors, getRelatedForm(elt), validate);
    }

    // include the element itself
    processInputValue(processed, formData, errors, elt, validate);

    // if a button or submit was clicked last, include its value
    if (internalData.lastButtonClicked || elt.tagName === 'BUTTON' ||
    (elt.tagName === 'INPUT' && getRawAttribute(elt, 'type') === 'submit')) {
      const button = internalData.lastButtonClicked || (/** @type HTMLInputElement|HTMLButtonElement */(elt));
      const name = getRawAttribute(button, 'name');
      addValueToFormData(name, button.value, priorityFormData);
    }

    // include any explicit includes
    const includes = findAttributeTargets(elt, 'hx-include');
    forEach(includes, function(node) {
      processInputValue(processed, formData, errors, asElement(node), validate);
      // if a non-form is included, include any input values within it
      if (!matches(node, 'form')) {
        forEach(asParentNode(node).querySelectorAll(INPUT_SELECTOR), function(descendant) {
          processInputValue(processed, formData, errors, descendant, validate);
        });
      }
    });

    // values from a <form> take precedence, overriding the regular values
    overrideFormData(formData, priorityFormData);

    return { errors, formData, values: formDataProxy(formData) }
  }

  /**
   * @param {string} returnStr
   * @param {string} name
   * @param {any} realValue
   * @returns {string}
   */
  function appendParam(returnStr, name, realValue) {
    if (returnStr !== '') {
      returnStr += '&';
    }
    if (String(realValue) === '[object Object]') {
      realValue = JSON.stringify(realValue);
    }
    const s = encodeURIComponent(realValue);
    returnStr += encodeURIComponent(name) + '=' + s;
    return returnStr
  }

  /**
   * @param {FormData|Object} values
   * @returns string
   */
  function urlEncode(values) {
    values = formDataFromObject(values);
    let returnStr = '';
    values.forEach(function(value, key) {
      returnStr = appendParam(returnStr, key, value);
    });
    return returnStr
  }

  //= ===================================================================
  // Ajax
  //= ===================================================================

  /**
 * @param {Element} elt
 * @param {Element} target
 * @param {string} prompt
 * @returns {HtmxHeaderSpecification}
 */
  function getHeaders(elt, target, prompt) {
    /** @type HtmxHeaderSpecification */
    const headers = {
      'HX-Request': 'true',
      'HX-Trigger': getRawAttribute(elt, 'id'),
      'HX-Trigger-Name': getRawAttribute(elt, 'name'),
      'HX-Target': getAttributeValue(target, 'id'),
      'HX-Current-URL': location.href
    };
    getValuesForElement(elt, 'hx-headers', false, headers);
    if (prompt !== undefined) {
      headers['HX-Prompt'] = prompt;
    }
    if (getInternalData(elt).boosted) {
      headers['HX-Boosted'] = 'true';
    }
    return headers
  }

  /**
 * filterValues takes an object containing form input values
 * and returns a new object that only contains keys that are
 * specified by the closest "hx-params" attribute
 * @param {FormData} inputValues
 * @param {Element} elt
 * @returns {FormData}
 */
  function filterValues(inputValues, elt) {
    const paramsValue = getClosestAttributeValue(elt, 'hx-params');
    if (paramsValue) {
      if (paramsValue === 'none') {
        return new FormData()
      } else if (paramsValue === '*') {
        return inputValues
      } else if (paramsValue.indexOf('not ') === 0) {
        forEach(paramsValue.slice(4).split(','), function(name) {
          name = name.trim();
          inputValues.delete(name);
        });
        return inputValues
      } else {
        const newValues = new FormData();
        forEach(paramsValue.split(','), function(name) {
          name = name.trim();
          if (inputValues.has(name)) {
            inputValues.getAll(name).forEach(function(value) { newValues.append(name, value); });
          }
        });
        return newValues
      }
    } else {
      return inputValues
    }
  }

  /**
   * @param {Element} elt
   * @return {boolean}
   */
  function isAnchorLink(elt) {
    return !!getRawAttribute(elt, 'href') && getRawAttribute(elt, 'href').indexOf('#') >= 0
  }

  /**
 * @param {Element} elt
 * @param {HtmxSwapStyle} [swapInfoOverride]
 * @returns {HtmxSwapSpecification}
 */
  function getSwapSpecification(elt, swapInfoOverride) {
    const swapInfo = swapInfoOverride || getClosestAttributeValue(elt, 'hx-swap');
    /** @type HtmxSwapSpecification */
    const swapSpec = {
      swapStyle: getInternalData(elt).boosted ? 'innerHTML' : htmx.config.defaultSwapStyle,
      swapDelay: htmx.config.defaultSwapDelay,
      settleDelay: htmx.config.defaultSettleDelay
    };
    if (htmx.config.scrollIntoViewOnBoost && getInternalData(elt).boosted && !isAnchorLink(elt)) {
      swapSpec.show = 'top';
    }
    if (swapInfo) {
      const split = splitOnWhitespace(swapInfo);
      if (split.length > 0) {
        for (let i = 0; i < split.length; i++) {
          const value = split[i];
          if (value.indexOf('swap:') === 0) {
            swapSpec.swapDelay = parseInterval(value.slice(5));
          } else if (value.indexOf('settle:') === 0) {
            swapSpec.settleDelay = parseInterval(value.slice(7));
          } else if (value.indexOf('transition:') === 0) {
            swapSpec.transition = value.slice(11) === 'true';
          } else if (value.indexOf('ignoreTitle:') === 0) {
            swapSpec.ignoreTitle = value.slice(12) === 'true';
          } else if (value.indexOf('scroll:') === 0) {
            const scrollSpec = value.slice(7);
            var splitSpec = scrollSpec.split(':');
            const scrollVal = splitSpec.pop();
            var selectorVal = splitSpec.length > 0 ? splitSpec.join(':') : null;
            // @ts-ignore
            swapSpec.scroll = scrollVal;
            swapSpec.scrollTarget = selectorVal;
          } else if (value.indexOf('show:') === 0) {
            const showSpec = value.slice(5);
            var splitSpec = showSpec.split(':');
            const showVal = splitSpec.pop();
            var selectorVal = splitSpec.length > 0 ? splitSpec.join(':') : null;
            swapSpec.show = showVal;
            swapSpec.showTarget = selectorVal;
          } else if (value.indexOf('focus-scroll:') === 0) {
            const focusScrollVal = value.slice('focus-scroll:'.length);
            swapSpec.focusScroll = focusScrollVal == 'true';
          } else if (i == 0) {
            swapSpec.swapStyle = value;
          } else {
            logError('Unknown modifier in hx-swap: ' + value);
          }
        }
      }
    }
    return swapSpec
  }

  /**
   * @param {Element} elt
   * @return {boolean}
   */
  function usesFormData(elt) {
    return getClosestAttributeValue(elt, 'hx-encoding') === 'multipart/form-data' ||
    (matches(elt, 'form') && getRawAttribute(elt, 'enctype') === 'multipart/form-data')
  }

  /**
   * @param {XMLHttpRequest} xhr
   * @param {Element} elt
   * @param {FormData} filteredParameters
   * @returns {*|string|null}
   */
  function encodeParamsForBody(xhr, elt, filteredParameters) {
    let encodedParameters = null;
    withExtensions(elt, function(extension) {
      if (encodedParameters == null) {
        encodedParameters = extension.encodeParameters(xhr, filteredParameters, elt);
      }
    });
    if (encodedParameters != null) {
      return encodedParameters
    } else {
      if (usesFormData(elt)) {
        // Force conversion to an actual FormData object in case filteredParameters is a formDataProxy
        // See https://github.com/bigskysoftware/htmx/issues/2317
        return overrideFormData(new FormData(), formDataFromObject(filteredParameters))
      } else {
        return urlEncode(filteredParameters)
      }
    }
  }

  /**
 *
 * @param {Element} target
 * @returns {HtmxSettleInfo}
 */
  function makeSettleInfo(target) {
    return { tasks: [], elts: [target] }
  }

  /**
   * @param {Element[]} content
   * @param {HtmxSwapSpecification} swapSpec
   */
  function updateScrollState(content, swapSpec) {
    const first = content[0];
    const last = content[content.length - 1];
    if (swapSpec.scroll) {
      var target = null;
      if (swapSpec.scrollTarget) {
        target = asElement(querySelectorExt(first, swapSpec.scrollTarget));
      }
      if (swapSpec.scroll === 'top' && (first || target)) {
        target = target || first;
        target.scrollTop = 0;
      }
      if (swapSpec.scroll === 'bottom' && (last || target)) {
        target = target || last;
        target.scrollTop = target.scrollHeight;
      }
      if (typeof swapSpec.scroll === 'number') {
        getWindow().setTimeout(function() {
          window.scrollTo(0, /** @type number */ (swapSpec.scroll));
        }, 0); // next 'tick', so browser has time to render layout
      }
    }
    if (swapSpec.show) {
      var target = null;
      if (swapSpec.showTarget) {
        let targetStr = swapSpec.showTarget;
        if (swapSpec.showTarget === 'window') {
          targetStr = 'body';
        }
        target = asElement(querySelectorExt(first, targetStr));
      }
      if (swapSpec.show === 'top' && (first || target)) {
        target = target || first;
        // @ts-ignore For some reason tsc doesn't recognize "instant" as a valid option for now
        target.scrollIntoView({ block: 'start', behavior: htmx.config.scrollBehavior });
      }
      if (swapSpec.show === 'bottom' && (last || target)) {
        target = target || last;
        // @ts-ignore For some reason tsc doesn't recognize "instant" as a valid option for now
        target.scrollIntoView({ block: 'end', behavior: htmx.config.scrollBehavior });
      }
    }
  }

  /**
 * @param {Element} elt
 * @param {string} attr
 * @param {boolean=} evalAsDefault
 * @param {Object=} values
 * @param {Event=} event
 * @returns {Object}
 */
  function getValuesForElement(elt, attr, evalAsDefault, values, event) {
    if (values == null) {
      values = {};
    }
    if (elt == null) {
      return values
    }
    const attributeValue = getAttributeValue(elt, attr);
    if (attributeValue) {
      let str = attributeValue.trim();
      let evaluateValue = evalAsDefault;
      if (str === 'unset') {
        return null
      }
      if (str.indexOf('javascript:') === 0) {
        str = str.slice(11);
        evaluateValue = true;
      } else if (str.indexOf('js:') === 0) {
        str = str.slice(3);
        evaluateValue = true;
      }
      if (str.indexOf('{') !== 0) {
        str = '{' + str + '}';
      }
      let varsValues;
      if (evaluateValue) {
        varsValues = maybeEval(elt, function() {
          if (event) {
            return Function('event', 'return (' + str + ')').call(elt, event)
          } else { // allow window.event to be accessible
            return Function('return (' + str + ')').call(elt)
          }
        }, {});
      } else {
        varsValues = parseJSON(str);
      }
      for (const key in varsValues) {
        if (varsValues.hasOwnProperty(key)) {
          if (values[key] == null) {
            values[key] = varsValues[key];
          }
        }
      }
    }
    return getValuesForElement(asElement(parentElt(elt)), attr, evalAsDefault, values, event)
  }

  /**
   * @param {EventTarget|string} elt
   * @param {() => any} toEval
   * @param {any=} defaultVal
   * @returns {any}
   */
  function maybeEval(elt, toEval, defaultVal) {
    if (htmx.config.allowEval) {
      return toEval()
    } else {
      triggerErrorEvent(elt, 'htmx:evalDisallowedError');
      return defaultVal
    }
  }

  /**
 * @param {Element} elt
 * @param {Event=} event
 * @param {*?=} expressionVars
 * @returns
 */
  function getHXVarsForElement(elt, event, expressionVars) {
    return getValuesForElement(elt, 'hx-vars', true, expressionVars, event)
  }

  /**
 * @param {Element} elt
 * @param {Event=} event
 * @param {*?=} expressionVars
 * @returns
 */
  function getHXValsForElement(elt, event, expressionVars) {
    return getValuesForElement(elt, 'hx-vals', false, expressionVars, event)
  }

  /**
 * @param {Element} elt
 * @param {Event=} event
 * @returns {FormData}
 */
  function getExpressionVars(elt, event) {
    return mergeObjects(getHXVarsForElement(elt, event), getHXValsForElement(elt, event))
  }

  /**
   * @param {XMLHttpRequest} xhr
   * @param {string} header
   * @param {string|null} headerValue
   */
  function safelySetHeaderValue(xhr, header, headerValue) {
    if (headerValue !== null) {
      try {
        xhr.setRequestHeader(header, headerValue);
      } catch (e) {
      // On an exception, try to set the header URI encoded instead
        xhr.setRequestHeader(header, encodeURIComponent(headerValue));
        xhr.setRequestHeader(header + '-URI-AutoEncoded', 'true');
      }
    }
  }

  /**
   * @param {XMLHttpRequest} xhr
   * @return {string}
   */
  function getPathFromResponse(xhr) {
    if (xhr.responseURL) {
      try {
        const url = new URL(xhr.responseURL);
        return url.pathname + url.search
      } catch (e) {
        triggerErrorEvent(getDocument().body, 'htmx:badResponseUrl', { url: xhr.responseURL });
      }
    }
  }

  /**
   * @param {XMLHttpRequest} xhr
   * @param {RegExp} regexp
   * @return {boolean}
   */
  function hasHeader(xhr, regexp) {
    return regexp.test(xhr.getAllResponseHeaders())
  }

  /**
   * Issues an htmx-style AJAX request
   *
   * @see https://htmx.org/api/#ajax
   *
   * @param {HttpVerb} verb
   * @param {string} path the URL path to make the AJAX
   * @param {Element|string|HtmxAjaxHelperContext} context the element to target (defaults to the **body**) | a selector for the target | a context object that contains any of the following
   * @return {Promise<void>} Promise that resolves immediately if no request is sent, or when the request is complete
   */
  function ajaxHelper(verb, path, context) {
    verb = (/** @type HttpVerb */(verb.toLowerCase()));
    if (context) {
      if (context instanceof Element || typeof context === 'string') {
        return issueAjaxRequest(verb, path, null, null, {
          targetOverride: resolveTarget(context) || DUMMY_ELT,
          returnPromise: true
        })
      } else {
        let resolvedTarget = resolveTarget(context.target);
        // If target is supplied but can't resolve OR source is supplied but both target and source can't be resolved
        // then use DUMMY_ELT to abort the request with htmx:targetError to avoid it replacing body by mistake
        if ((context.target && !resolvedTarget) || (context.source && !resolvedTarget && !resolveTarget(context.source))) {
          resolvedTarget = DUMMY_ELT;
        }
        return issueAjaxRequest(verb, path, resolveTarget(context.source), context.event,
          {
            handler: context.handler,
            headers: context.headers,
            values: context.values,
            targetOverride: resolvedTarget,
            swapOverride: context.swap,
            select: context.select,
            returnPromise: true,
            push: context.push,
            replace: context.replace,
            selectOOB: context.selectOOB
          })
      }
    } else {
      return issueAjaxRequest(verb, path, null, null, {
        returnPromise: true
      })
    }
  }

  /**
   * @param {Element} elt
   * @return {Element[]}
   */
  function hierarchyForElt(elt) {
    const arr = [];
    while (elt) {
      arr.push(elt);
      elt = elt.parentElement;
    }
    return arr
  }

  /**
   * @param {Element} elt
   * @param {string} path
   * @param {HtmxRequestConfig} requestConfig
   * @return {boolean}
   */
  function verifyPath(elt, path, requestConfig) {
    const url = new URL(path, location.protocol !== 'about:' ? location.href : window.origin);
    const origin = location.protocol !== 'about:' ? location.origin : window.origin;
    const sameHost = origin === url.origin;

    if (htmx.config.selfRequestsOnly) {
      if (!sameHost) {
        return false
      }
    }
    return triggerEvent(elt, 'htmx:validateUrl', mergeObjects({ url, sameHost }, requestConfig))
  }

  /**
   * @param {Object|FormData} obj
   * @return {FormData}
   */
  function formDataFromObject(obj) {
    if (obj instanceof FormData) return obj
    const formData = new FormData();
    for (const key in obj) {
      if (obj.hasOwnProperty(key)) {
        if (obj[key] && typeof obj[key].forEach === 'function') {
          obj[key].forEach(function(v) { formData.append(key, v); });
        } else if (typeof obj[key] === 'object' && !(obj[key] instanceof Blob)) {
          formData.append(key, JSON.stringify(obj[key]));
        } else {
          formData.append(key, obj[key]);
        }
      }
    }
    return formData
  }

  /**
   * @param {FormData} formData
   * @param {string} name
   * @param {Array} array
   * @returns {Array}
   */
  function formDataArrayProxy(formData, name, array) {
    // mutating the array should mutate the underlying form data
    return new Proxy(array, {
      get: function(target, key) {
        if (typeof key === 'number') return target[key]
        if (key === 'length') return target.length
        if (key === 'push') {
          return function(value) {
            target.push(value);
            formData.append(name, value);
          }
        }
        if (typeof target[key] === 'function') {
          return function() {
            target[key].apply(target, arguments);
            formData.delete(name);
            target.forEach(function(v) { formData.append(name, v); });
          }
        }

        if (target[key] && target[key].length === 1) {
          return target[key][0]
        } else {
          return target[key]
        }
      },
      set: function(target, index, value) {
        target[index] = value;
        formData.delete(name);
        target.forEach(function(v) { formData.append(name, v); });
        return true
      }
    })
  }

  /**
   * @param {FormData} formData
   * @returns {Object}
   */
  function formDataProxy(formData) {
    return new Proxy(formData, {
      get: function(target, name) {
        if (typeof name === 'symbol') {
          // Forward symbol calls to the FormData itself directly
          const result = Reflect.get(target, name);
          // Wrap in function with apply to correctly bind the FormData context, as a direct call would result in an illegal invocation error
          if (typeof result === 'function') {
            return function() {
              return result.apply(formData, arguments)
            }
          } else {
            return result
          }
        }
        if (name === 'toJSON') {
          // Support JSON.stringify call on proxy
          return () => Object.fromEntries(formData)
        }
        if (name in target) {
          // Wrap in function with apply to correctly bind the FormData context, as a direct call would result in an illegal invocation error
          if (typeof target[name] === 'function') {
            return function() {
              return formData[name].apply(formData, arguments)
            }
          }
        }
        const array = formData.getAll(name);
        // Those 2 undefined & single value returns are for retro-compatibility as we weren't using FormData before
        if (array.length === 0) {
          return undefined
        } else if (array.length === 1) {
          return array[0]
        } else {
          return formDataArrayProxy(target, name, array)
        }
      },
      set: function(target, name, value) {
        if (typeof name !== 'string') {
          return false
        }
        target.delete(name);
        if (value && typeof value.forEach === 'function') {
          value.forEach(function(v) { target.append(name, v); });
        } else if (typeof value === 'object' && !(value instanceof Blob)) {
          target.append(name, JSON.stringify(value));
        } else {
          target.append(name, value);
        }
        return true
      },
      deleteProperty: function(target, name) {
        if (typeof name === 'string') {
          target.delete(name);
        }
        return true
      },
      // Support Object.assign call from proxy
      ownKeys: function(target) {
        return Reflect.ownKeys(Object.fromEntries(target))
      },
      getOwnPropertyDescriptor: function(target, prop) {
        return Reflect.getOwnPropertyDescriptor(Object.fromEntries(target), prop)
      }
    })
  }

  /**
   * @param {HttpVerb} verb
   * @param {string} path
   * @param {Element} elt
   * @param {Event} event
   * @param {HtmxAjaxEtc} [etc]
   * @param {boolean} [confirmed]
   * @return {Promise<void>}
   */
  function issueAjaxRequest(verb, path, elt, event, etc, confirmed) {
    let resolve = null;
    let reject = null;
    etc = etc != null ? etc : {};
    if (etc.returnPromise && typeof Promise !== 'undefined') {
      var promise = new Promise(function(_resolve, _reject) {
        resolve = _resolve;
        reject = _reject;
      });
    }
    if (elt == null) {
      elt = getDocument().body;
    }
    const responseHandler = etc.handler || handleAjaxResponse;
    const select = etc.select || null;

    if (!bodyContains(elt)) {
    // do not issue requests for elements removed from the DOM
      maybeCall(resolve);
      return promise
    }
    const target = etc.targetOverride || asElement(getTarget(elt));
    if (target == null || target == DUMMY_ELT) {
      triggerErrorEvent(elt, 'htmx:targetError', { target: getClosestAttributeValue(elt, 'hx-target') });
      maybeCall(reject);
      return promise
    }

    let eltData = getInternalData(elt);
    const submitter = eltData.lastButtonClicked;

    if (submitter) {
      const buttonPath = getRawAttribute(submitter, 'formaction');
      if (buttonPath != null) {
        path = buttonPath;
      }

      const buttonVerb = getRawAttribute(submitter, 'formmethod');
      if (buttonVerb != null) {
        if (VERBS.includes(buttonVerb.toLowerCase())) {
          verb = (/** @type HttpVerb */(buttonVerb));
        } else {
          maybeCall(resolve);
          return promise
        }
      }
    }

    const confirmQuestion = getClosestAttributeValue(elt, 'hx-confirm');
    // allow event-based confirmation w/ a callback
    if (confirmed === undefined) {
      const issueRequest = function(skipConfirmation) {
        return issueAjaxRequest(verb, path, elt, event, etc, !!skipConfirmation)
      };
      const confirmDetails = { target, elt, path, verb, triggeringEvent: event, etc, issueRequest, question: confirmQuestion };
      if (triggerEvent(elt, 'htmx:confirm', confirmDetails) === false) {
        maybeCall(resolve);
        return promise
      }
    }

    let syncElt = elt;
    let syncStrategy = getClosestAttributeValue(elt, 'hx-sync');
    let queueStrategy = null;
    let abortable = false;
    if (syncStrategy) {
      const syncStrings = syncStrategy.split(':');
      const selector = syncStrings[0].trim();
      if (selector === 'this') {
        syncElt = findThisElement(elt, 'hx-sync');
      } else {
        syncElt = asElement(querySelectorExt(elt, selector));
      }
      // default to the drop strategy
      syncStrategy = (syncStrings[1] || 'drop').trim();
      eltData = getInternalData(syncElt);
      if (syncStrategy === 'drop' && eltData.xhr && eltData.abortable !== true) {
        maybeCall(resolve);
        return promise
      } else if (syncStrategy === 'abort') {
        if (eltData.xhr) {
          maybeCall(resolve);
          return promise
        } else {
          abortable = true;
        }
      } else if (syncStrategy === 'replace') {
        triggerEvent(syncElt, 'htmx:abort'); // abort the current request and continue
      } else if (syncStrategy.indexOf('queue') === 0) {
        const queueStrArray = syncStrategy.split(' ');
        queueStrategy = (queueStrArray[1] || 'last').trim();
      }
    }

    if (eltData.xhr) {
      if (eltData.abortable) {
        triggerEvent(syncElt, 'htmx:abort'); // abort the current request and continue
      } else {
        if (queueStrategy == null) {
          if (event) {
            const eventData = getInternalData(event);
            if (eventData && eventData.triggerSpec && eventData.triggerSpec.queue) {
              queueStrategy = eventData.triggerSpec.queue;
            }
          }
          if (queueStrategy == null) {
            queueStrategy = 'last';
          }
        }
        if (eltData.queuedRequests == null) {
          eltData.queuedRequests = [];
        }
        if (queueStrategy === 'first' && eltData.queuedRequests.length === 0) {
          eltData.queuedRequests.push(function() {
            issueAjaxRequest(verb, path, elt, event, etc);
          });
        } else if (queueStrategy === 'all') {
          eltData.queuedRequests.push(function() {
            issueAjaxRequest(verb, path, elt, event, etc);
          });
        } else if (queueStrategy === 'last') {
          eltData.queuedRequests = []; // dump existing queue
          eltData.queuedRequests.push(function() {
            issueAjaxRequest(verb, path, elt, event, etc);
          });
        }
        maybeCall(resolve);
        return promise
      }
    }

    const xhr = new XMLHttpRequest();
    eltData.xhr = xhr;
    eltData.abortable = abortable;
    const endRequestLock = function() {
      eltData.xhr = null;
      eltData.abortable = false;
      if (eltData.queuedRequests != null &&
      eltData.queuedRequests.length > 0) {
        const queuedRequest = eltData.queuedRequests.shift();
        queuedRequest();
      }
    };
    const promptQuestion = getClosestAttributeValue(elt, 'hx-prompt');
    if (promptQuestion) {
      var promptResponse = prompt(promptQuestion);
      // prompt returns null if cancelled and empty string if accepted with no entry
      if (promptResponse === null ||
      !triggerEvent(elt, 'htmx:prompt', { prompt: promptResponse, target })) {
        maybeCall(resolve);
        endRequestLock();
        return promise
      }
    }

    if (confirmQuestion && !confirmed) {
      if (!confirm(confirmQuestion)) {
        maybeCall(resolve);
        endRequestLock();
        return promise
      }
    }

    let headers = getHeaders(elt, target, promptResponse);

    if (verb !== 'get' && !usesFormData(elt)) {
      headers['Content-Type'] = 'application/x-www-form-urlencoded';
    }

    if (etc.headers) {
      headers = mergeObjects(headers, etc.headers);
    }
    const results = getInputValues(elt, verb);
    let errors = results.errors;
    const rawFormData = results.formData;
    if (etc.values) {
      overrideFormData(rawFormData, formDataFromObject(etc.values));
    }
    const expressionVars = formDataFromObject(getExpressionVars(elt, event));
    const allFormData = overrideFormData(rawFormData, expressionVars);
    let filteredFormData = filterValues(allFormData, elt);

    if (htmx.config.getCacheBusterParam && verb === 'get') {
      filteredFormData.set('org.htmx.cache-buster', getRawAttribute(target, 'id') || 'true');
    }

    // behavior of anchors w/ empty href is to use the current URL
    if (path == null || path === '') {
      path = location.href;
    }

    /**
     * @type {Object}
     * @property {boolean} [credentials]
     * @property {number} [timeout]
     * @property {boolean} [noHeaders]
     */
    const requestAttrValues = getValuesForElement(elt, 'hx-request');

    const eltIsBoosted = getInternalData(elt).boosted;

    let useUrlParams = htmx.config.methodsThatUseUrlParams.indexOf(verb) >= 0;

    /** @type HtmxRequestConfig */
    const requestConfig = {
      boosted: eltIsBoosted,
      useUrlParams,
      formData: filteredFormData,
      parameters: formDataProxy(filteredFormData),
      unfilteredFormData: allFormData,
      unfilteredParameters: formDataProxy(allFormData),
      headers,
      elt,
      target,
      verb,
      errors,
      withCredentials: etc.credentials || requestAttrValues.credentials || htmx.config.withCredentials,
      timeout: etc.timeout || requestAttrValues.timeout || htmx.config.timeout,
      path,
      triggeringEvent: event
    };

    if (!triggerEvent(elt, 'htmx:configRequest', requestConfig)) {
      maybeCall(resolve);
      endRequestLock();
      return promise
    }

    // copy out in case the object was overwritten
    path = requestConfig.path;
    verb = requestConfig.verb;
    headers = requestConfig.headers;
    filteredFormData = formDataFromObject(requestConfig.parameters);
    errors = requestConfig.errors;
    useUrlParams = requestConfig.useUrlParams;

    if (errors && errors.length > 0) {
      triggerEvent(elt, 'htmx:validation:halted', requestConfig);
      maybeCall(resolve);
      endRequestLock();
      return promise
    }

    const splitPath = path.split('#');
    const pathNoAnchor = splitPath[0];
    const anchor = splitPath[1];

    let finalPath = path;
    if (useUrlParams) {
      finalPath = pathNoAnchor;
      const hasValues = !filteredFormData.keys().next().done;
      if (hasValues) {
        if (finalPath.indexOf('?') < 0) {
          finalPath += '?';
        } else {
          finalPath += '&';
        }
        finalPath += urlEncode(filteredFormData);
        if (anchor) {
          finalPath += '#' + anchor;
        }
      }
    }

    if (!verifyPath(elt, finalPath, requestConfig)) {
      triggerErrorEvent(elt, 'htmx:invalidPath', requestConfig);
      maybeCall(reject);
      endRequestLock();
      return promise
    }

    xhr.open(verb.toUpperCase(), finalPath, true);
    xhr.overrideMimeType('text/html');
    xhr.withCredentials = requestConfig.withCredentials;
    xhr.timeout = requestConfig.timeout;

    // request headers
    if (requestAttrValues.noHeaders) ; else {
      for (const header in headers) {
        if (headers.hasOwnProperty(header)) {
          const headerValue = headers[header];
          safelySetHeaderValue(xhr, header, headerValue);
        }
      }
    }

    /** @type {HtmxResponseInfo} */
    const responseInfo = {
      xhr,
      target,
      requestConfig,
      etc,
      boosted: eltIsBoosted,
      select,
      pathInfo: {
        requestPath: path,
        finalRequestPath: finalPath,
        responsePath: null,
        anchor
      }
    };

    xhr.onload = function() {
      try {
        const hierarchy = hierarchyForElt(elt);
        responseInfo.pathInfo.responsePath = getPathFromResponse(xhr);
        responseHandler(elt, responseInfo);
        if (responseInfo.keepIndicators !== true) {
          removeRequestIndicators(indicators, disableElts);
        }
        triggerEvent(elt, 'htmx:afterRequest', responseInfo);
        triggerEvent(elt, 'htmx:afterOnLoad', responseInfo);
        // if the body no longer contains the element, trigger the event on the closest parent
        // remaining in the DOM
        if (!bodyContains(elt)) {
          let secondaryTriggerElt = null;
          while (hierarchy.length > 0 && secondaryTriggerElt == null) {
            const parentEltInHierarchy = hierarchy.shift();
            if (bodyContains(parentEltInHierarchy)) {
              secondaryTriggerElt = parentEltInHierarchy;
            }
          }
          if (secondaryTriggerElt) {
            triggerEvent(secondaryTriggerElt, 'htmx:afterRequest', responseInfo);
            triggerEvent(secondaryTriggerElt, 'htmx:afterOnLoad', responseInfo);
          }
        }
        maybeCall(resolve);
      } catch (e) {
        triggerErrorEvent(elt, 'htmx:onLoadError', mergeObjects({ error: e }, responseInfo));
        throw e
      } finally {
        endRequestLock();
      }
    };
    xhr.onerror = function() {
      removeRequestIndicators(indicators, disableElts);
      triggerErrorEvent(elt, 'htmx:afterRequest', responseInfo);
      triggerErrorEvent(elt, 'htmx:sendError', responseInfo);
      maybeCall(reject);
      endRequestLock();
    };
    xhr.onabort = function() {
      removeRequestIndicators(indicators, disableElts);
      triggerErrorEvent(elt, 'htmx:afterRequest', responseInfo);
      triggerErrorEvent(elt, 'htmx:sendAbort', responseInfo);
      maybeCall(reject);
      endRequestLock();
    };
    xhr.ontimeout = function() {
      removeRequestIndicators(indicators, disableElts);
      triggerErrorEvent(elt, 'htmx:afterRequest', responseInfo);
      triggerErrorEvent(elt, 'htmx:timeout', responseInfo);
      maybeCall(reject);
      endRequestLock();
    };
    if (!triggerEvent(elt, 'htmx:beforeRequest', responseInfo)) {
      maybeCall(resolve);
      endRequestLock();
      return promise
    }
    var indicators = addRequestIndicatorClasses(elt);
    var disableElts = disableElements(elt);

    forEach(['loadstart', 'loadend', 'progress', 'abort'], function(eventName) {
      forEach([xhr, xhr.upload], function(target) {
        target.addEventListener(eventName, function(event) {
          triggerEvent(elt, 'htmx:xhr:' + eventName, {
            lengthComputable: event.lengthComputable,
            loaded: event.loaded,
            total: event.total
          });
        });
      });
    });
    triggerEvent(elt, 'htmx:beforeSend', responseInfo);
    const params = useUrlParams ? null : encodeParamsForBody(xhr, elt, filteredFormData);
    xhr.send(params);
    return promise
  }

  /**
   * @typedef {Object} HtmxHistoryUpdate
   * @property {string|null} [type]
   * @property {string|null} [path]
   */

  /**
   * @param {Element} elt
   * @param {HtmxResponseInfo} responseInfo
   * @return {HtmxHistoryUpdate}
   */
  function determineHistoryUpdates(elt, responseInfo) {
    const xhr = responseInfo.xhr;

    //= ==========================================
    // First consult response headers
    //= ==========================================
    let pathFromHeaders = null;
    let typeFromHeaders = null;
    if (hasHeader(xhr, /HX-Push:/i)) {
      pathFromHeaders = xhr.getResponseHeader('HX-Push');
      typeFromHeaders = 'push';
    } else if (hasHeader(xhr, /HX-Push-Url:/i)) {
      pathFromHeaders = xhr.getResponseHeader('HX-Push-Url');
      typeFromHeaders = 'push';
    } else if (hasHeader(xhr, /HX-Replace-Url:/i)) {
      pathFromHeaders = xhr.getResponseHeader('HX-Replace-Url');
      typeFromHeaders = 'replace';
    }

    // if there was a response header, that has priority
    if (pathFromHeaders) {
      if (pathFromHeaders === 'false') {
        return {}
      } else {
        return {
          type: typeFromHeaders,
          path: pathFromHeaders
        }
      }
    }

    //= ==========================================
    // Next resolve via DOM values
    //= ==========================================
    const requestPath = responseInfo.pathInfo.finalRequestPath;
    const responsePath = responseInfo.pathInfo.responsePath;

    const pushUrl = responseInfo.etc.push || getClosestAttributeValue(elt, 'hx-push-url');
    const replaceUrl = responseInfo.etc.replace || getClosestAttributeValue(elt, 'hx-replace-url');
    const elementIsBoosted = getInternalData(elt).boosted;

    let saveType = null;
    let path = null;

    if (pushUrl) {
      saveType = 'push';
      path = pushUrl;
    } else if (replaceUrl) {
      saveType = 'replace';
      path = replaceUrl;
    } else if (elementIsBoosted) {
      saveType = 'push';
      path = responsePath || requestPath; // if there is no response path, go with the original request path
    }

    if (path) {
    // false indicates no push, return empty object
      if (path === 'false') {
        return {}
      }

      // true indicates we want to follow wherever the server ended up sending us
      if (path === 'true') {
        path = responsePath || requestPath; // if there is no response path, go with the original request path
      }

      // restore any anchor associated with the request
      if (responseInfo.pathInfo.anchor && path.indexOf('#') === -1) {
        path = path + '#' + responseInfo.pathInfo.anchor;
      }

      return {
        type: saveType,
        path
      }
    } else {
      return {}
    }
  }

  /**
   * @param {HtmxResponseHandlingConfig} responseHandlingConfig
   * @param {number} status
   * @return {boolean}
   */
  function codeMatches(responseHandlingConfig, status) {
    var regExp = new RegExp(responseHandlingConfig.code);
    return regExp.test(status.toString(10))
  }

  /**
   * @param {XMLHttpRequest} xhr
   * @return {HtmxResponseHandlingConfig}
   */
  function resolveResponseHandling(xhr) {
    for (var i = 0; i < htmx.config.responseHandling.length; i++) {
      /** @type HtmxResponseHandlingConfig */
      var responseHandlingElement = htmx.config.responseHandling[i];
      if (codeMatches(responseHandlingElement, xhr.status)) {
        return responseHandlingElement
      }
    }
    // no matches, return no swap
    return {
      swap: false
    }
  }

  /**
   * @param {string} title
   */
  function handleTitle(title) {
    if (title) {
      const titleElt = find('title');
      if (titleElt) {
        titleElt.textContent = title;
      } else {
        window.document.title = title;
      }
    }
  }

  /**
   * Resove the Retarget selector and throw if not found
   * @param {Element} elt
   * @param {String} target
   * @returns {Element}
   */
  function resolveRetarget(elt, target) {
    if (target === 'this') {
      return elt
    }
    const resolvedTarget = asElement(querySelectorExt(elt, target));
    if (resolvedTarget == null) {
      triggerErrorEvent(elt, 'htmx:targetError', { target });
      throw new Error(`Invalid re-target ${target}`)
    }
    return resolvedTarget
  }

  /**
   * @param {Element} elt
   * @param {HtmxResponseInfo} responseInfo
   */
  function handleAjaxResponse(elt, responseInfo) {
    const xhr = responseInfo.xhr;
    let target = responseInfo.target;
    const etc = responseInfo.etc;
    const responseInfoSelect = responseInfo.select;

    if (!triggerEvent(elt, 'htmx:beforeOnLoad', responseInfo)) return

    if (hasHeader(xhr, /HX-Trigger:/i)) {
      handleTriggerHeader(xhr, 'HX-Trigger', elt);
    }

    if (hasHeader(xhr, /HX-Location:/i)) {
      let redirectPath = xhr.getResponseHeader('HX-Location');
      /** @type {HtmxAjaxHelperContext&{path?:string}} */
      var redirectSwapSpec = {};
      if (redirectPath.indexOf('{') === 0) {
        redirectSwapSpec = parseJSON(redirectPath);
        // what's the best way to throw an error if the user didn't include this
        redirectPath = redirectSwapSpec.path;
        delete redirectSwapSpec.path;
      }
      redirectSwapSpec.push = redirectSwapSpec.push || 'true';
      ajaxHelper('get', redirectPath, redirectSwapSpec);
      return
    }

    const shouldRefresh = hasHeader(xhr, /HX-Refresh:/i) && xhr.getResponseHeader('HX-Refresh') === 'true';

    if (hasHeader(xhr, /HX-Redirect:/i)) {
      responseInfo.keepIndicators = true;
      htmx.location.href = xhr.getResponseHeader('HX-Redirect');
      shouldRefresh && htmx.location.reload();
      return
    }

    if (shouldRefresh) {
      responseInfo.keepIndicators = true;
      htmx.location.reload();
      return
    }

    const historyUpdate = determineHistoryUpdates(elt, responseInfo);

    const responseHandling = resolveResponseHandling(xhr);
    const shouldSwap = responseHandling.swap;
    let isError = !!responseHandling.error;
    let ignoreTitle = htmx.config.ignoreTitle || responseHandling.ignoreTitle;
    let selectOverride = responseHandling.select;
    if (responseHandling.target) {
      responseInfo.target = resolveRetarget(elt, responseHandling.target);
    }
    var swapOverride = etc.swapOverride;
    if (swapOverride == null && responseHandling.swapOverride) {
      swapOverride = responseHandling.swapOverride;
    }

    // response headers override response handling config
    if (hasHeader(xhr, /HX-Retarget:/i)) {
      responseInfo.target = resolveRetarget(elt, xhr.getResponseHeader('HX-Retarget'));
    }

    if (hasHeader(xhr, /HX-Reswap:/i)) {
      swapOverride = xhr.getResponseHeader('HX-Reswap');
    }

    var serverResponse = xhr.response;
    /** @type HtmxBeforeSwapDetails */
    var beforeSwapDetails = mergeObjects({
      shouldSwap,
      serverResponse,
      isError,
      ignoreTitle,
      selectOverride,
      swapOverride
    }, responseInfo);

    if (responseHandling.event && !triggerEvent(target, responseHandling.event, beforeSwapDetails)) return

    if (!triggerEvent(target, 'htmx:beforeSwap', beforeSwapDetails)) return

    target = beforeSwapDetails.target; // allow re-targeting
    serverResponse = beforeSwapDetails.serverResponse; // allow updating content
    isError = beforeSwapDetails.isError; // allow updating error
    ignoreTitle = beforeSwapDetails.ignoreTitle; // allow updating ignoring title
    selectOverride = beforeSwapDetails.selectOverride; // allow updating select override
    swapOverride = beforeSwapDetails.swapOverride; // allow updating swap override

    responseInfo.target = target; // Make updated target available to response events
    responseInfo.failed = isError; // Make failed property available to response events
    responseInfo.successful = !isError; // Make successful property available to response events

    if (beforeSwapDetails.shouldSwap) {
      if (xhr.status === 286) {
        cancelPolling(elt);
      }

      withExtensions(elt, function(extension) {
        serverResponse = extension.transformResponse(serverResponse, xhr, elt);
      });

      // Save current page if there will be a history update
      if (historyUpdate.type) {
        saveCurrentPageToHistory();
      }

      var swapSpec = getSwapSpecification(elt, swapOverride);

      if (!swapSpec.hasOwnProperty('ignoreTitle')) {
        swapSpec.ignoreTitle = ignoreTitle;
      }

      target.classList.add(htmx.config.swappingClass);

      if (responseInfoSelect) {
        selectOverride = responseInfoSelect;
      }

      if (hasHeader(xhr, /HX-Reselect:/i)) {
        selectOverride = xhr.getResponseHeader('HX-Reselect');
      }

      const selectOOB = etc.selectOOB || getClosestAttributeValue(elt, 'hx-select-oob');
      const select = getClosestAttributeValue(elt, 'hx-select');

      swap(target, serverResponse, swapSpec, {
        select: selectOverride === 'unset' ? null : selectOverride || select,
        selectOOB,
        eventInfo: responseInfo,
        anchor: responseInfo.pathInfo.anchor,
        contextElement: elt,
        afterSwapCallback: function() {
          if (hasHeader(xhr, /HX-Trigger-After-Swap:/i)) {
            let finalElt = elt;
            if (!bodyContains(elt)) {
              finalElt = getDocument().body;
            }
            handleTriggerHeader(xhr, 'HX-Trigger-After-Swap', finalElt);
          }
        },
        afterSettleCallback: function() {
          if (hasHeader(xhr, /HX-Trigger-After-Settle:/i)) {
            let finalElt = elt;
            if (!bodyContains(elt)) {
              finalElt = getDocument().body;
            }
            handleTriggerHeader(xhr, 'HX-Trigger-After-Settle', finalElt);
          }
        },
        beforeSwapCallback: function() {
          // if we need to save history, do so, before swapping so that relative resources have the correct base URL
          if (historyUpdate.type) {
            triggerEvent(getDocument().body, 'htmx:beforeHistoryUpdate', mergeObjects({ history: historyUpdate }, responseInfo));
            if (historyUpdate.type === 'push') {
              pushUrlIntoHistory(historyUpdate.path);
              triggerEvent(getDocument().body, 'htmx:pushedIntoHistory', { path: historyUpdate.path });
            } else {
              replaceUrlInHistory(historyUpdate.path);
              triggerEvent(getDocument().body, 'htmx:replacedInHistory', { path: historyUpdate.path });
            }
          }
        }
      });
    }
    if (isError) {
      triggerErrorEvent(elt, 'htmx:responseError', mergeObjects({ error: 'Response Status Error Code ' + xhr.status + ' from ' + responseInfo.pathInfo.requestPath }, responseInfo));
    }
  }

  //= ===================================================================
  // Extensions API
  //= ===================================================================

  /** @type {Object<string, HtmxExtension>} */
  const extensions = {};

  /**
   * extensionBase defines the default functions for all extensions.
   * @returns {HtmxExtension}
   */
  function extensionBase() {
    return {
      init: function(api) { return null },
      getSelectors: function() { return null },
      onEvent: function(name, evt) { return true },
      transformResponse: function(text, xhr, elt) { return text },
      isInlineSwap: function(swapStyle) { return false },
      handleSwap: function(swapStyle, target, fragment, settleInfo) { return false },
      encodeParameters: function(xhr, parameters, elt) { return null }
    }
  }

  /**
   * defineExtension initializes the extension and adds it to the htmx registry
   *
   * @see https://htmx.org/api/#defineExtension
   *
   * @param {string} name the extension name
   * @param {Partial<HtmxExtension>} extension the extension definition
   */
  function defineExtension(name, extension) {
    if (extension.init) {
      extension.init(internalAPI);
    }
    extensions[name] = mergeObjects(extensionBase(), extension);
  }

  /**
   * removeExtension removes an extension from the htmx registry
   *
   * @see https://htmx.org/api/#removeExtension
   *
   * @param {string} name
   */
  function removeExtension(name) {
    delete extensions[name];
  }

  /**
   * getExtensions searches up the DOM tree to return all extensions that can be applied to a given element
   *
   * @param {Element} elt
   * @param {HtmxExtension[]=} extensionsToReturn
   * @param {string[]=} extensionsToIgnore
   * @returns {HtmxExtension[]}
   */
  function getExtensions(elt, extensionsToReturn, extensionsToIgnore) {
    if (extensionsToReturn == undefined) {
      extensionsToReturn = [];
    }
    if (elt == undefined) {
      return extensionsToReturn
    }
    if (extensionsToIgnore == undefined) {
      extensionsToIgnore = [];
    }
    const extensionsForElement = getAttributeValue(elt, 'hx-ext');
    if (extensionsForElement) {
      forEach(extensionsForElement.split(','), function(extensionName) {
        extensionName = extensionName.replace(/ /g, '');
        if (extensionName.slice(0, 7) == 'ignore:') {
          extensionsToIgnore.push(extensionName.slice(7));
          return
        }
        if (extensionsToIgnore.indexOf(extensionName) < 0) {
          const extension = extensions[extensionName];
          if (extension && extensionsToReturn.indexOf(extension) < 0) {
            extensionsToReturn.push(extension);
          }
        }
      });
    }
    return getExtensions(asElement(parentElt(elt)), extensionsToReturn, extensionsToIgnore)
  }

  //= ===================================================================
  // Initialization
  //= ===================================================================
  var isReady = false;
  getDocument().addEventListener('DOMContentLoaded', function() {
    isReady = true;
  });

  /**
   * Execute a function now if DOMContentLoaded has fired, otherwise listen for it.
   *
   * This function uses isReady because there is no reliable way to ask the browser whether
   * the DOMContentLoaded event has already been fired; there's a gap between DOMContentLoaded
   * firing and readystate=complete.
   */
  function ready(fn) {
    // Checking readyState here is a failsafe in case the htmx script tag entered the DOM by
    // some means other than the initial page load.
    if (isReady || getDocument().readyState === 'complete') {
      fn();
    } else {
      getDocument().addEventListener('DOMContentLoaded', fn);
    }
  }

  function insertIndicatorStyles() {
    if (htmx.config.includeIndicatorStyles !== false) {
      const nonceAttribute = htmx.config.inlineStyleNonce ? ` nonce="${htmx.config.inlineStyleNonce}"` : '';
      const indicator = htmx.config.indicatorClass;
      const request = htmx.config.requestClass;
      getDocument().head.insertAdjacentHTML('beforeend',
        `<style${nonceAttribute}>` +
        `.${indicator}{opacity:0;visibility: hidden} ` +
        `.${request} .${indicator}, .${request}.${indicator}{opacity:1;visibility: visible;transition: opacity 200ms ease-in}` +
        '</style>'
      );
    }
  }

  function getMetaConfig() {
    /** @type HTMLMetaElement */
    const element = getDocument().querySelector('meta[name="htmx-config"]');
    if (element) {
      return parseJSON(element.content)
    } else {
      return null
    }
  }

  function mergeMetaConfig() {
    const metaConfig = getMetaConfig();
    if (metaConfig) {
      htmx.config = mergeObjects(htmx.config, metaConfig);
    }
  }

  // initialize the document
  ready(function() {
    mergeMetaConfig();
    insertIndicatorStyles();
    let body = getDocument().body;
    processNode(body);
    const restoredElts = getDocument().querySelectorAll(
      "[hx-trigger='restored'],[data-hx-trigger='restored']"
    );
    body.addEventListener('htmx:abort', function(evt) {
      const target = (/** @type {CustomEvent} */(evt)).detail.elt || evt.target;
      const internalData = getInternalData(target);
      if (internalData && internalData.xhr) {
        internalData.xhr.abort();
      }
    });
    /** @type {(ev: PopStateEvent) => any} */
    const originalPopstate = window.onpopstate ? window.onpopstate.bind(window) : null;
    /** @type {(ev: PopStateEvent) => any} */
    window.onpopstate = function(event) {
      if (event.state && event.state.htmx) {
        restoreHistory();
        forEach(restoredElts, function(elt) {
          triggerEvent(elt, 'htmx:restored', {
            document: getDocument(),
            triggerEvent
          });
        });
      } else {
        if (originalPopstate) {
          originalPopstate(event);
        }
      }
    };
    getWindow().setTimeout(function() {
      triggerEvent(body, 'htmx:load', {}); // give ready handlers a chance to load up before firing this event
      body = null; // kill reference for gc
    }, 0);
  });

  return htmx
})();

var scannerLoader = {};

/**
 * Scanner Loader - Lazy loads html5-qrcode only when scanning is initiated
 * This reduces initial bundle size by ~90KB gzipped (html5-qrcode library)
 */

var hasRequiredScannerLoader;

function requireScannerLoader () {
	if (hasRequiredScannerLoader) return scannerLoader;
	hasRequiredScannerLoader = 1;
	let Html5QrcodeModule = null;
	let scannerModuleLoading = false;
	let scannerModuleLoadPromise = null;

	/**
	 * Lazy load Html5Qrcode library
	 * @returns {Promise<Object>} Html5Qrcode module
	 */
	async function loadScannerModule () {
	  if (Html5QrcodeModule) {
	    return Html5QrcodeModule
	  }

	  if (scannerModuleLoading) {
	    return scannerModuleLoadPromise
	  }

	  scannerModuleLoading = true;
	  scannerModuleLoadPromise = import('./index.js').then(module => {
	    Html5QrcodeModule = module;

	    // Make globally available for scanner.js
	    window.Html5Qrcode = module.Html5Qrcode;
	    window.Html5QrcodeSupportedFormats = {
	      QR_CODE: 0,
	      AZTEC: 1,
	      CODABAR: 2,
	      CODE_39: 3,
	      CODE_93: 4,
	      CODE_128: 5,
	      DATA_MATRIX: 6,
	      MAXICODE: 7,
	      ITF: 8,
	      EAN_13: 9,
	      EAN_8: 10,
	      PDF_417: 11,
	      UPC_A: 12,
	      UPC_E: 13,
	      UPC_EAN_EXTENSION: 14,
	      ITF_14: 15
	    };

	    return module
	  });

	  return scannerModuleLoadPromise
	}

	// Barcode format mapping
	const BARCODE_TYPE_MAPPING = {
	  QR_CODE: 'QR',
	  AZTEC: 'AZTEC',
	  DATA_MATRIX: 'DATAMATRIX',
	  PDF_417: 'PDF417',
	  MAXICODE: 'MAXICODE',
	  CODE_128: 'CODE128',
	  CODE_39: 'CODE39',
	  CODE_93: 'CODE93',
	  CODABAR: 'CODABAR',
	  EAN_13: 'EAN13',
	  EAN_8: 'EAN8',
	  UPC_A: 'UPCA',
	  UPC_E: 'UPCE',
	  ITF: 'ITF',
	  ITF_14: 'ITF14'
	};

	/**
	 * Create a barcode scanner with lazy loading
	 * @param {Object} config - Scanner configuration
	 * @param {Array<number>} formats - Supported barcode formats
	 * @returns {Object} Alpine.js component data
	 */
	function createBarcodeScanner (config, formats = null) {
	  return {
	    isNewMerchant: false,
	    [config.fieldName]: config.initialValue,
	    barcodeType: '',
	    scanning: false,
	    scanMessage: config.defaultMessage,
	    html5QrCode: null,

	    getSupportedFormats () {
	      if (formats && formats.length > 0) {
	        return formats
	      }

	      // All supported formats if none specified
	      const Html5QrcodeSupportedFormats = window.Html5QrcodeSupportedFormats;
	      return [
	        Html5QrcodeSupportedFormats.QR_CODE,
	        Html5QrcodeSupportedFormats.AZTEC,
	        Html5QrcodeSupportedFormats.DATA_MATRIX,
	        Html5QrcodeSupportedFormats.PDF_417,
	        Html5QrcodeSupportedFormats.MAXICODE,
	        Html5QrcodeSupportedFormats.CODE_128,
	        Html5QrcodeSupportedFormats.CODE_39,
	        Html5QrcodeSupportedFormats.CODE_93,
	        Html5QrcodeSupportedFormats.CODABAR,
	        Html5QrcodeSupportedFormats.EAN_13,
	        Html5QrcodeSupportedFormats.EAN_8,
	        Html5QrcodeSupportedFormats.UPC_A,
	        Html5QrcodeSupportedFormats.UPC_E,
	        Html5QrcodeSupportedFormats.ITF,
	        Html5QrcodeSupportedFormats.ITF_14
	      ]
	    },

	    async startScanning () {
	      this.scanning = true;
	      this.scanMessage = 'Scanner wird geladen...';

	      try {
	        // Lazy load scanner module
	        await loadScannerModule();

	        const { Html5Qrcode } = window;
	        if (!Html5Qrcode) {
	          throw new Error('Html5Qrcode not loaded')
	        }

	        this.scanMessage = 'Kamera wird gestartet...';
	        this.html5QrCode = new Html5Qrcode('qr-reader');

	        await this.html5QrCode.start(
	          { facingMode: 'environment' },
	          {
	            fps: 10,
	            qrbox: { width: 300, height: 300 },
	            formatsToSupport: this.getSupportedFormats(),
	            aspectRatio: 1.0,
	            disableFlip: false
	          },
	          (decodedText, decodedResult) => {
	            this.onScanSuccess(decodedText, decodedResult);
	          },
	          (errorMessage) => {
	            // Optional: Debug scan attempts
	          }
	        );

	        this.scanMessage = 'Halte den Barcode in den Rahmen';
	      } catch (err) {
	        console.error('Scanner error:', err);

	        let message = 'Kamera-Zugriff fehlgeschlagen';
	        if (err.name === 'NotAllowedError') {
	          message = 'Kamera-Berechtigung verweigert';
	        } else if (err.name === 'NotFoundError') {
	          message = 'Keine Kamera gefunden';
	        } else if (err.name === 'NotReadableError') {
	          message = 'Kamera wird bereits verwendet';
	        }

	        this.scanMessage = message;
	        setTimeout(() => { this.scanning = false; }, 3000);
	      }
	    },

	    onScanSuccess (decodedText, decodedResult) {
	      const formatName = decodedResult.result.format.formatName;
	      const internalType = BARCODE_TYPE_MAPPING[formatName] || formatName;

	      if (this.isISBN10(decodedText)) {
	        this[config.fieldName] = this.convertISBN10toISBN13(decodedText);
	        this.barcodeType = 'ISBN13';
	        this.scanMessage = 'ISBN-10 erkannt (zu ISBN-13 konvertiert)';
	      } else {
	        this[config.fieldName] = decodedText;
	        this.barcodeType = internalType;
	        this.scanMessage = `${formatName} erkannt!`;
	      }

	      this.$nextTick(() => {
	        this.updateBarcodeTypeDropdown();
	      });

	      setTimeout(() => this.stopScanning(), 1000);
	    },

	    updateBarcodeTypeDropdown () {
	      const dropdown = document.querySelector('select[name="barcode_type"]');
	      if (dropdown && this.barcodeType) {
	        const option = dropdown.querySelector(`option[value="${this.barcodeType}"]`);
	        if (option) {
	          dropdown.value = this.barcodeType;
	          dropdown.dispatchEvent(new Event('change', { bubbles: true }));
	        }
	      }
	    },

	    stopScanning () {
	      if (this.html5QrCode) {
	        this.html5QrCode.stop().then(() => {
	          this.html5QrCode.clear();
	          this.scanning = false;
	          this.scanMessage = config.defaultMessage;
	        }).catch(err => {
	          console.error('Stop error:', err);
	          this.scanning = false;
	          this.scanMessage = config.defaultMessage;
	        });
	      }
	    },

	    isISBN10 (code) {
	      return /^\d{9}[\dX]$/.test(code)
	    },

	    convertISBN10toISBN13 (isbn10) {
	      const isbn9 = isbn10.substring(0, 9);
	      const isbn12 = '978' + isbn9;

	      let sum = 0;
	      for (let i = 0; i < 12; i++) {
	        sum += parseInt(isbn12[i]) * (i % 2 === 0 ? 1 : 3);
	      }
	      const checkDigit = (10 - (sum % 10)) % 10;

	      return isbn12 + checkDigit
	    }
	  }
	}

	// Export scanner factory functions
	window.cardForm = function (initialCardNumber = '') {
	  const Html5QrcodeSupportedFormats = window.Html5QrcodeSupportedFormats || {};
	  return createBarcodeScanner(
	    {
	      fieldName: 'cardNumber',
	      initialValue: initialCardNumber,
	      defaultMessage: 'Halte die Karte in den Rahmen'},
	    [
	      Html5QrcodeSupportedFormats.CODE_128 || 5,
	      Html5QrcodeSupportedFormats.EAN_13 || 9,
	      Html5QrcodeSupportedFormats.EAN_8 || 10,
	      Html5QrcodeSupportedFormats.QR_CODE || 0,
	      Html5QrcodeSupportedFormats.CODE_39 || 3,
	      Html5QrcodeSupportedFormats.UPC_A || 12
	    ]
	  )
	};

	window.voucherForm = function (initialCode = '') {
	  return createBarcodeScanner(
	    {
	      fieldName: 'voucherCode',
	      initialValue: initialCode,
	      defaultMessage: 'Halte den Gutschein-Code in den Rahmen'},
	    null // All formats
	  )
	};

	window.giftCardForm = function (initialCardNumber = '') {
	  const Html5QrcodeSupportedFormats = window.Html5QrcodeSupportedFormats || {};
	  const scanner = createBarcodeScanner(
	    {
	      fieldName: 'cardNumber',
	      initialValue: initialCardNumber,
	      defaultMessage: 'Halte die Kartennummer in den Rahmen'},
	    [
	      Html5QrcodeSupportedFormats.CODE_128 || 5,
	      Html5QrcodeSupportedFormats.EAN_13 || 9,
	      Html5QrcodeSupportedFormats.QR_CODE || 0,
	      Html5QrcodeSupportedFormats.PDF_417 || 11,
	      Html5QrcodeSupportedFormats.UPC_A || 12,
	      Html5QrcodeSupportedFormats.CODE_39 || 3
	    ]
	  );

	  scanner.merchantId = '';
	  return scanner
	};

	window.emailAutocomplete = function () {
	  return {
	    email: '',
	    suggestions: [],
	    showSuggestions: false,
	    loading: false,
	    selectedIndex: -1,

	    async fetchSuggestions () {
	      if (this.email.length < 1) {
	        this.suggestions = [];
	        this.showSuggestions = false;
	        return
	      }

	      this.loading = true;

	      try {
	        const response = await fetch(`/api/shared-users?q=${encodeURIComponent(this.email)}`);
	        const data = await response.json();
	        this.suggestions = data || [];
	        this.showSuggestions = this.suggestions.length > 0;
	        this.selectedIndex = -1;
	      } catch (error) {
	        console.error('Failed to fetch suggestions:', error);
	        this.suggestions = [];
	        this.showSuggestions = false;
	      } finally {
	        this.loading = false;
	      }
	    },

	    selectSuggestion (suggestion) {
	      this.email = suggestion.email;
	      this.showSuggestions = false;
	      this.suggestions = [];
	      this.selectedIndex = -1;
	    },

	    handleKeydown (event) {
	      if (!this.showSuggestions || this.suggestions.length === 0) return

	      if (event.key === 'ArrowDown') {
	        event.preventDefault();
	        this.selectedIndex = Math.min(this.selectedIndex + 1, this.suggestions.length - 1);
	      } else if (event.key === 'ArrowUp') {
	        event.preventDefault();
	        this.selectedIndex = Math.max(this.selectedIndex - 1, -1);
	      } else if (event.key === 'Enter' && this.selectedIndex >= 0) {
	        event.preventDefault();
	        this.selectSuggestion(this.suggestions[this.selectedIndex]);
	      } else if (event.key === 'Escape') {
	        this.showSuggestions = false;
	        this.selectedIndex = -1;
	      }
	    },

	    hideSuggestions () {
	      setTimeout(() => {
	        this.showSuggestions = false;
	        this.selectedIndex = -1;
	      }, 200);
	    }
	  }
	};
	return scannerLoader;
}

requireScannerLoader();

// Toast notification helper
function showToast (message, type = 'info') {
  const toast = document.createElement('div');
  toast.className = `fixed bottom-4 right-4 px-6 py-3 rounded-lg shadow-lg text-white font-medium z-50 transition-all duration-300 ${
    type === 'success'
? 'bg-green-600'
    : type === 'warning'
? 'bg-yellow-600'
    : type === 'error'
? 'bg-red-600'
    : 'bg-blue-600'
  }`;
  toast.textContent = message;
  toast.style.opacity = '0';
  toast.style.transform = 'translateY(20px)';

  document.body.appendChild(toast);

  // Animate in
  setTimeout(() => {
    toast.style.opacity = '1';
    toast.style.transform = 'translateY(0)';
  }, 10);

  // Animate out and remove
  setTimeout(() => {
    toast.style.opacity = '0';
    toast.style.transform = 'translateY(20px)';
    setTimeout(() => toast.remove(), 300);
  }, 3000);
}

/**
 * Reliable connection check
 * Based on: https://dev.to/maxmonteil/is-your-app-online-here-s-how-to-reliably-know-in-just-10-lines-of-js-guide-3in7
 *
 * navigator.onLine is unreliable:
 * - false = always reliable (you are offline)
 * - true = unreliable (only means "connected to network", not "has internet")
 */
async function checkOnlineStatus () {
  // If navigator says offline, trust it (false is always reliable)
  if (!navigator.onLine) {
    return false
  }

  // If navigator says online, verify with real fetch
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 2000);

    // Use health endpoint with GET (HEAD might redirect)
    // Add timestamp to prevent cache
    const response = await fetch(`/health?t=${Date.now()}`, {
      method: 'GET',
      cache: 'no-store',
      signal: controller.signal
    });

    clearTimeout(timeoutId);
    return response.ok
  } catch {
    // Network error or timeout = offline
    return false
  }
}

// Offline Detection - Alpine.js Store
function initOfflineStore (Alpine) {
  console.log('[Alpine] Init - creating offline store');

  // Create store with initial state
  Alpine.store('offline', {
    isOnline: navigator.onLine,
    checking: false
  });

  // Immediately verify actual connection on page load
  checkOnlineStatus().then(isOnline => {
    console.log('[Offline] Initial check - isOnline:', isOnline);
    Alpine.store('offline').isOnline = isOnline;
  });

  // Listen to browser online/offline events
  window.addEventListener('online', async () => {
    console.log('[Offline] Browser online event');
    // Verify with real check
    const isOnline = await checkOnlineStatus();
    Alpine.store('offline').isOnline = isOnline;
    if (isOnline) {
      showToast('✅ Verbindung wiederhergestellt', 'success');
    }
  });

  window.addEventListener('offline', () => {
    console.log('[Offline] Browser offline event');
    Alpine.store('offline').isOnline = false;
  });

  // Periodic check every 5 seconds (backup for missed events)
  setInterval(async () => {
    const isOnline = await checkOnlineStatus();
    const storeOnline = Alpine.store('offline').isOnline;

    if (isOnline !== storeOnline) {
      console.log('[Offline] Periodic check detected change:', isOnline);
      Alpine.store('offline').isOnline = isOnline;
    }
  }, 5000);

  // Add manual check method to store
  const store = Alpine.store('offline');
  store.checkConnection = async function () {
    this.checking = true;
    console.log('[Offline] Manual check...');

    const isOnline = await checkOnlineStatus();
    this.isOnline = isOnline;
    this.checking = false;

    if (isOnline) {
      console.log('[Offline] Server reachable - reloading...');
      setTimeout(() => window.location.reload(), 300);
    } else {
      console.log('[Offline] Still offline');
    }
  };
}

// Orientation-based fullscreen barcode overlay
// When on mobile and switching to landscape, show the barcode fullscreen
// so it's easier to scan at a shop.

function initOrientationStore (Alpine) {
  Alpine.store('barcode', {
    fullscreen: false,
    barcodeHtml: ''
  });

  // Only activate on touch devices (mobile)
  if (!('ontouchstart' in window)) {
    return
  }

  const store = Alpine.store('barcode');

  function getBarcodeSectionHtml () {
    const section = document.getElementById('barcode-section');
    if (!section) return ''
    return section.innerHTML
  }

  function handleOrientationChange () {
    const isLandscape =
      window.screen.orientation?.type?.includes('landscape') ||
      window.innerWidth > window.innerHeight;

    const section = document.getElementById('barcode-section');
    if (!section) {
      store.fullscreen = false;
      return
    }

    if (isLandscape) {
      store.barcodeHtml = getBarcodeSectionHtml();
      store.fullscreen = true;
    } else {
      store.fullscreen = false;
      store.barcodeHtml = '';
    }
  }

  // Listen to orientation changes
  window.addEventListener('orientationchange', () => {
    // Small delay so innerWidth/innerHeight are updated
    setTimeout(handleOrientationChange, 100);
  });

  // Fallback: resize event (some browsers fire this instead)
  let resizeTimer = null;
  window.addEventListener('resize', () => {
    if (resizeTimer) clearTimeout(resizeTimer);
    resizeTimer = setTimeout(handleOrientationChange, 150);
  });

  // Close on click/touch anywhere on the overlay
  store.close = function () {
    this.fullscreen = false;
    this.barcodeHtml = '';
  };
}

// Precache - Cache all detail pages for offline access
let precacheInitialized = false;
const cachedUrls = new Set();

function setupPrecaching () {
  // Only run once
  if (precacheInitialized || !('serviceWorker' in navigator)) {
    return
  }
  precacheInitialized = true;

  console.log('[Precache] Initializing...');

  // Helper to cache a URL (deduplicated)
  function cacheUrl (url) {
    if (cachedUrls.has(url)) {
      return
    }
    cachedUrls.add(url);

    fetch(url, {
      method: 'GET',
      credentials: 'same-origin',
      cache: 'no-store'
    }).catch(() => {
      // Silently fail, remove from set to retry later
      cachedUrls.delete(url);
    });
  }

  // Fetch detail page and cache its barcode images
  async function cacheDetailPageWithBarcodes (detailUrl) {
    try {
      const response = await fetch(detailUrl, {
        method: 'GET',
        credentials: 'same-origin',
        cache: 'no-store'
      });

      if (!response.ok) return

      cachedUrls.add(detailUrl);

      // Parse HTML to find barcode images
      const html = await response.text();
      const parser = new DOMParser();
      const doc = parser.parseFromString(html, 'text/html');

      // Find barcode images
      const barcodeImages = doc.querySelectorAll('img[src^="/barcode"]');
      barcodeImages.forEach(img => {
        const barcodeUrl = img.getAttribute('src');
        if (barcodeUrl) {
          console.log('[Precache] Caching barcode:', barcodeUrl);
          cacheUrl(barcodeUrl);
        }
      });
    } catch (err) {
      console.warn('[Precache] Failed to cache detail page:', detailUrl, err);
    }
  }

  // Fetch and parse list pages to find all detail links
  async function cacheListPageAndDetails (listUrl) {
    console.log('[Precache] Fetching list page:', listUrl);
    try {
      // First cache the list page itself
      const response = await fetch(listUrl, {
        method: 'GET',
        credentials: 'same-origin',
        cache: 'no-store'
      });

      if (!response.ok) {
        console.warn('[Precache] Failed to fetch list page:', listUrl, response.status);
        return
      }

      cachedUrls.add(listUrl);
      console.log('[Precache] Cached list page:', listUrl);

      // Parse HTML to find detail page links
      const html = await response.text();
      const parser = new DOMParser();
      const doc = parser.parseFromString(html, 'text/html');

      // Find all detail page links in this list page
      const links = doc.querySelectorAll('a[href^="/cards/"], a[href^="/vouchers/"], a[href^="/gift-cards/"]');
      console.log(`[Precache] Found ${links.length} links in ${listUrl}`);

      links.forEach(link => {
        const url = link.getAttribute('href');
        if (/^\/(?:cards|vouchers|gift-cards)\/[0-9a-f-]{36}/.test(url)) {
          console.log('[Precache] Caching detail page with barcodes:', url);
          // Cache detail page AND its barcodes
          cacheDetailPageWithBarcodes(url);
        }
      });
    } catch (err) {
      console.warn('[Precache] Failed to cache list page:', listUrl, err);
    }
  }

  // Cache dashboard and all three list pages with their details
  async function cacheAllPages () {
    // First cache dashboard
    cacheUrl('/');

    // Then cache all list pages and their details
    const listPages = ['/cards', '/vouchers', '/gift-cards'];
    for (const page of listPages) {
      await cacheListPageAndDetails(page);
    }

    console.log('[Precache] Initial caching complete');
  }

  // Start caching
  cacheAllPages();

  // Function to scan for detail pages in current DOM
  function scanForDetailPages () {
    const links = document.querySelectorAll('a[href^="/cards/"], a[href^="/vouchers/"], a[href^="/gift-cards/"]');

    links.forEach(link => {
      const url = link.getAttribute('href');
      if (/^\/(?:cards|vouchers|gift-cards)\/[0-9a-f-]{36}/.test(url)) {
        cacheUrl(url);
      }
    });
  }

  // Initial scan after page load
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', scanForDetailPages);
  } else {
    scanForDetailPages();
  }

  // Re-scan after HTMX swaps
  if (typeof htmx !== 'undefined') {
    document.body.addEventListener('htmx:afterSwap', () => {
      setTimeout(scanForDetailPages, 100);
    });
  }
}

// static/js/src/app.js - Main entry point for bundled JS


// Make Alpine global
window.Alpine = module_default;

// Make HTMX global
window.htmx = htmx$1;

// Initialize offline store before starting Alpine
initOfflineStore(module_default);

// Initialize orientation store for landscape barcode fullscreen
initOrientationStore(module_default);

// Setup precaching for detail pages
setupPrecaching();

// Start Alpine
module_default.start();
//# sourceMappingURL=bundle.js.map
