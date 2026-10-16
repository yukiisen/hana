let wasm;

const heap = new Array(128).fill(undefined);

heap.push(undefined, null, true, false);

function getObject(idx) { return heap[idx]; }

let heap_next = heap.length;

function addHeapObject(obj) {
    if (heap_next === heap.length) heap.push(heap.length + 1);
    const idx = heap_next;
    heap_next = heap[idx];

    heap[idx] = obj;
    return idx;
}

function handleError(f, args) {
    try {
        return f.apply(this, args);
    } catch (e) {
        wasm.__wbindgen_export_0(addHeapObject(e));
    }
}

const cachedTextDecoder = (typeof TextDecoder !== 'undefined' ? new TextDecoder('utf-8', { ignoreBOM: true, fatal: true }) : { decode: () => { throw Error('TextDecoder not available') } } );

if (typeof TextDecoder !== 'undefined') { cachedTextDecoder.decode(); };

let cachedUint8ArrayMemory0 = null;

function getUint8ArrayMemory0() {
    if (cachedUint8ArrayMemory0 === null || cachedUint8ArrayMemory0.byteLength === 0) {
        cachedUint8ArrayMemory0 = new Uint8Array(wasm.memory.buffer);
    }
    return cachedUint8ArrayMemory0;
}

function getStringFromWasm0(ptr, len) {
    ptr = ptr >>> 0;
    return cachedTextDecoder.decode(getUint8ArrayMemory0().subarray(ptr, ptr + len));
}

function dropObject(idx) {
    if (idx < 132) return;
    heap[idx] = heap_next;
    heap_next = idx;
}

function takeObject(idx) {
    const ret = getObject(idx);
    dropObject(idx);
    return ret;
}

function debugString(val) {
    // primitive types
    const type = typeof val;
    if (type == 'number' || type == 'boolean' || val == null) {
        return  `${val}`;
    }
    if (type == 'string') {
        return `"${val}"`;
    }
    if (type == 'symbol') {
        const description = val.description;
        if (description == null) {
            return 'Symbol';
        } else {
            return `Symbol(${description})`;
        }
    }
    if (type == 'function') {
        const name = val.name;
        if (typeof name == 'string' && name.length > 0) {
            return `Function(${name})`;
        } else {
            return 'Function';
        }
    }
    // objects
    if (Array.isArray(val)) {
        const length = val.length;
        let debug = '[';
        if (length > 0) {
            debug += debugString(val[0]);
        }
        for(let i = 1; i < length; i++) {
            debug += ', ' + debugString(val[i]);
        }
        debug += ']';
        return debug;
    }
    // Test for built-in
    const builtInMatches = /\[object ([^\]]+)\]/.exec(toString.call(val));
    let className;
    if (builtInMatches && builtInMatches.length > 1) {
        className = builtInMatches[1];
    } else {
        // Failed to match the standard '[object ClassName]'
        return toString.call(val);
    }
    if (className == 'Object') {
        // we're a user defined class or Object
        // JSON.stringify avoids problems with cycles, and is generally much
        // easier than looping through ownProperties of `val`.
        try {
            return 'Object(' + JSON.stringify(val) + ')';
        } catch (_) {
            return 'Object';
        }
    }
    // errors
    if (val instanceof Error) {
        return `${val.name}: ${val.message}\n${val.stack}`;
    }
    // TODO we could test for more things here, like `Set`s and `Map`s.
    return className;
}

let WASM_VECTOR_LEN = 0;

const cachedTextEncoder = (typeof TextEncoder !== 'undefined' ? new TextEncoder('utf-8') : { encode: () => { throw Error('TextEncoder not available') } } );

const encodeString = (typeof cachedTextEncoder.encodeInto === 'function'
    ? function (arg, view) {
    return cachedTextEncoder.encodeInto(arg, view);
}
    : function (arg, view) {
    const buf = cachedTextEncoder.encode(arg);
    view.set(buf);
    return {
        read: arg.length,
        written: buf.length
    };
});

function passStringToWasm0(arg, malloc, realloc) {

    if (realloc === undefined) {
        const buf = cachedTextEncoder.encode(arg);
        const ptr = malloc(buf.length, 1) >>> 0;
        getUint8ArrayMemory0().subarray(ptr, ptr + buf.length).set(buf);
        WASM_VECTOR_LEN = buf.length;
        return ptr;
    }

    let len = arg.length;
    let ptr = malloc(len, 1) >>> 0;

    const mem = getUint8ArrayMemory0();

    let offset = 0;

    for (; offset < len; offset++) {
        const code = arg.charCodeAt(offset);
        if (code > 0x7F) break;
        mem[ptr + offset] = code;
    }

    if (offset !== len) {
        if (offset !== 0) {
            arg = arg.slice(offset);
        }
        ptr = realloc(ptr, len, len = offset + arg.length * 3, 1) >>> 0;
        const view = getUint8ArrayMemory0().subarray(ptr + offset, ptr + len);
        const ret = encodeString(arg, view);

        offset += ret.written;
        ptr = realloc(ptr, len, offset, 1) >>> 0;
    }

    WASM_VECTOR_LEN = offset;
    return ptr;
}

let cachedDataViewMemory0 = null;

function getDataViewMemory0() {
    if (cachedDataViewMemory0 === null || cachedDataViewMemory0.buffer.detached === true || (cachedDataViewMemory0.buffer.detached === undefined && cachedDataViewMemory0.buffer !== wasm.memory.buffer)) {
        cachedDataViewMemory0 = new DataView(wasm.memory.buffer);
    }
    return cachedDataViewMemory0;
}

function isLikeNone(x) {
    return x === undefined || x === null;
}

function _assertClass(instance, klass) {
    if (!(instance instanceof klass)) {
        throw new Error(`expected instance of ${klass.name}`);
    }
}

let stack_pointer = 128;

function addBorrowedObject(obj) {
    if (stack_pointer == 1) throw new Error('out of js stack');
    heap[--stack_pointer] = obj;
    return stack_pointer;
}

function getArrayJsValueFromWasm0(ptr, len) {
    ptr = ptr >>> 0;
    const mem = getDataViewMemory0();
    const result = [];
    for (let i = ptr; i < ptr + 4 * len; i += 4) {
        result.push(takeObject(mem.getUint32(i, true)));
    }
    return result;
}

let cachedFloat64ArrayMemory0 = null;

function getFloat64ArrayMemory0() {
    if (cachedFloat64ArrayMemory0 === null || cachedFloat64ArrayMemory0.byteLength === 0) {
        cachedFloat64ArrayMemory0 = new Float64Array(wasm.memory.buffer);
    }
    return cachedFloat64ArrayMemory0;
}

function getArrayF64FromWasm0(ptr, len) {
    ptr = ptr >>> 0;
    return getFloat64ArrayMemory0().subarray(ptr / 8, ptr / 8 + len);
}
/**
 * @enum {0 | 1 | 2 | 3}
 */
export const GameMode = Object.freeze({
    Osu: 0, "0": "Osu",
    Taiko: 1, "1": "Taiko",
    Catch: 2, "2": "Catch",
    Mania: 3, "3": "Mania",
});
/**
 * A specific implementation of hitresult generation.
 * @enum {0 | 1}
 */
export const HitResultGenerator = Object.freeze({
    /**
     * Prioritize generating hitresults quickly.
     */
    Fast: 0, "0": "Fast",
    /**
     * Find the hitresults that match the given accuracy the closest.
     */
    Closest: 1, "1": "Closest",
});
/**
 * While generating remaining hitresults, decide how they should be distributed.
 * @enum {0 | 1}
 */
export const HitResultPriority = Object.freeze({
    /**
     * Prioritize good hitresults over bad ones
     */
    BestCase: 0, "0": "BestCase",
    /**
     * Prioritize bad hitresults over good ones
     */
    WorstCase: 1, "1": "WorstCase",
});

const BeatmapFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_beatmap_free(ptr >>> 0, 1));
/**
 * All beatmap data that is relevant for difficulty and performance
 * calculation.
 *
 * It is recommended to call the method `Beatmap.free` on instances that are
 * no longer in use to avoid the risk of leaking memory.
 */
export class Beatmap {

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        BeatmapFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_beatmap_free(ptr, 0);
    }
    /**
     * @returns {boolean}
     */
    get isConvert() {
        const ret = wasm.beatmap_isConvert(this.__wbg_ptr);
        return ret !== 0;
    }
    /**
     * @returns {number}
     */
    get nSpinners() {
        const ret = wasm.beatmap_nSpinners(this.__wbg_ptr);
        return ret >>> 0;
    }
    /**
     * Check whether hitobjects appear too suspicious for further calculation.
     *
     * Sometimes a beatmap isn't created for gameplay but rather to test
     * the limits of osu! itself. Difficulty- and/or performance calculation
     * should likely be avoided on these maps due to potential performance
     * issues.
     * @returns {boolean}
     */
    isSuspicious() {
        const ret = wasm.beatmap_isSuspicious(this.__wbg_ptr);
        return ret !== 0;
    }
    /**
     * @returns {number}
     */
    get stackLeniency() {
        const ret = wasm.beatmap_stackLeniency(this.__wbg_ptr);
        return ret;
    }
    /**
     * @returns {number}
     */
    get sliderTickRate() {
        const ret = wasm.beatmap_sliderTickRate(this.__wbg_ptr);
        return ret;
    }
    /**
     * @returns {number}
     */
    get sliderMultiplier() {
        const ret = wasm.beatmap_sliderMultiplier(this.__wbg_ptr);
        return ret;
    }
    /**
     * @returns {number}
     */
    get ar() {
        const ret = wasm.beatmap_ar(this.__wbg_ptr);
        return ret;
    }
    /**
     * @returns {number}
     */
    get cs() {
        const ret = wasm.beatmap_cs(this.__wbg_ptr);
        return ret;
    }
    /**
     * @returns {number}
     */
    get hp() {
        const ret = wasm.beatmap_hp(this.__wbg_ptr);
        return ret;
    }
    /**
     * @returns {number}
     */
    get od() {
        const ret = wasm.beatmap_od(this.__wbg_ptr);
        return ret;
    }
    /**
     * @returns {number}
     */
    get bpm() {
        const ret = wasm.beatmap_bpm(this.__wbg_ptr);
        return ret;
    }
    /**
     * Create a new beatmap instance by parsing an `.osu` file's content.
     * @throws Throws an error if decoding the map failed
     * @param {BeatmapContent} args
     */
    constructor(args) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.beatmap_new(retptr, addBorrowedObject(args));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            var r2 = getDataViewMemory0().getInt32(retptr + 4 * 2, true);
            if (r2) {
                throw takeObject(r1);
            }
            this.__wbg_ptr = r0 >>> 0;
            BeatmapFinalization.register(this, this.__wbg_ptr, this);
            return this;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
            heap[stack_pointer++] = undefined;
        }
    }
    /**
     * @returns {GameMode}
     */
    get mode() {
        const ret = wasm.beatmap_mode(this.__wbg_ptr);
        return ret;
    }
    /**
     * Convert a beatmap to a specific mode.
     * @throws Throws an error if conversion fails or mods are invalid
     * @param {GameMode} mode
     * @param {Object | null} [mods]
     */
    convert(mode, mods) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.beatmap_convert(retptr, this.__wbg_ptr, mode, isLikeNone(mods) ? 0 : addHeapObject(mods));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            if (r1) {
                throw takeObject(r0);
            }
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @returns {number}
     */
    get nHolds() {
        const ret = wasm.beatmap_nHolds(this.__wbg_ptr);
        return ret >>> 0;
    }
    /**
     * @returns {number}
     */
    get version() {
        const ret = wasm.beatmap_version(this.__wbg_ptr);
        return ret;
    }
    /**
     * @returns {number}
     */
    get nBreaks() {
        const ret = wasm.beatmap_nBreaks(this.__wbg_ptr);
        return ret >>> 0;
    }
    /**
     * @returns {number}
     */
    get nCircles() {
        const ret = wasm.beatmap_nCircles(this.__wbg_ptr);
        return ret >>> 0;
    }
    /**
     * @returns {number}
     */
    get nObjects() {
        const ret = wasm.beatmap_nObjects(this.__wbg_ptr);
        return ret >>> 0;
    }
    /**
     * @returns {number}
     */
    get nSliders() {
        const ret = wasm.beatmap_nSliders(this.__wbg_ptr);
        return ret >>> 0;
    }
}

const BeatmapAttributesFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_beatmapattributes_free(ptr >>> 0, 1));

export class BeatmapAttributes {

    static __wrap(ptr) {
        ptr = ptr >>> 0;
        const obj = Object.create(BeatmapAttributes.prototype);
        obj.__wbg_ptr = ptr;
        BeatmapAttributesFinalization.register(obj, obj.__wbg_ptr, obj);
        return obj;
    }

    toJSON() {
        return {
            ar: this.ar,
            baseAr: this.baseAr,
            od: this.od,
            baseOd: this.baseOd,
            cs: this.cs,
            hp: this.hp,
            clockRate: this.clockRate,
            arHitWindow: this.arHitWindow,
            odPerfectHitWindow: this.odPerfectHitWindow,
            odGreatHitWindow: this.odGreatHitWindow,
            odGoodHitWindow: this.odGoodHitWindow,
            odOkHitWindow: this.odOkHitWindow,
            odMehHitWindow: this.odMehHitWindow,
        };
    }

    toString() {
        return JSON.stringify(this);
    }

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        BeatmapAttributesFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_beatmapattributes_free(ptr, 0);
    }
    /**
     * The approach rate.
     * @returns {number}
     */
    get ar() {
        const ret = wasm.__wbg_get_beatmapattributes_ar(this.__wbg_ptr);
        return ret;
    }
    /**
     * The base approach rate without considering clock rate.
     * @returns {number}
     */
    get baseAr() {
        const ret = wasm.__wbg_get_beatmapattributes_baseAr(this.__wbg_ptr);
        return ret;
    }
    /**
     * The overall difficulty.
     * @returns {number}
     */
    get od() {
        const ret = wasm.__wbg_get_beatmapattributes_od(this.__wbg_ptr);
        return ret;
    }
    /**
     * The base overall difficulty without considering clock rate.
     * @returns {number}
     */
    get baseOd() {
        const ret = wasm.__wbg_get_beatmapattributes_baseOd(this.__wbg_ptr);
        return ret;
    }
    /**
     * The circle size.
     * @returns {number}
     */
    get cs() {
        const ret = wasm.__wbg_get_beatmapattributes_cs(this.__wbg_ptr);
        return ret;
    }
    /**
     * The health drain rate
     * @returns {number}
     */
    get hp() {
        const ret = wasm.__wbg_get_beatmapattributes_hp(this.__wbg_ptr);
        return ret;
    }
    /**
     * The clock rate with respect to mods.
     * @returns {number}
     */
    get clockRate() {
        const ret = wasm.__wbg_get_beatmapattributes_clockRate(this.__wbg_ptr);
        return ret;
    }
    /**
     * Hit window for approach rate i.e. TimePreempt in milliseconds.
     *
     * Only available for osu!standard and osu!catch.
     * @returns {number | undefined}
     */
    get arHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_arHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Perfect hit window for overall difficulty i.e. time to hit "Perfect" in
     * milliseconds.
     *
     * Only available for osu!mania.
     * @returns {number | undefined}
     */
    get odPerfectHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odPerfectHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Great hit window for overall difficulty i.e. time to hit a 300 ("Great")
     * in milliseconds.
     *
     * Only available for osu!standard, osu!taiko, and osu!mania.
     * @returns {number | undefined}
     */
    get odGreatHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odGreatHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Good hit window for overall difficulty i.e. time to hit a 200 ("Good")
     * in milliseconds.
     *
     * Only available for osu!mania.
     * @returns {number | undefined}
     */
    get odGoodHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odGoodHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Ok hit window for overall difficulty i.e. time to hit a 100 ("Ok") in
     * milliseconds.
     *
     * Only available for osu!standard, osu!taiko, and osu!mania.
     * @returns {number | undefined}
     */
    get odOkHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odOkHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Meh hit window for overall difficulty i.e. time to hit a 50 ("Meh") in
     * milliseconds.
     *
     * Only available for osu!standard and osu!mania.
     * @returns {number | undefined}
     */
    get odMehHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odMehHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
}

const BeatmapAttributesBuilderFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_beatmapattributesbuilder_free(ptr >>> 0, 1));

export class BeatmapAttributesBuilder {

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        BeatmapAttributesBuilderFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_beatmapattributesbuilder_free(ptr, 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedAr(fixed) {
        wasm.beatmapattributesbuilder_set_fixed_ar(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedCs(fixed) {
        wasm.beatmapattributesbuilder_set_fixed_cs(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedHp(fixed) {
        wasm.beatmapattributesbuilder_set_fixed_hp(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedOd(fixed) {
        wasm.beatmapattributesbuilder_set_fixed_od(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {number | null} [clock_rate]
     */
    set clockRate(clock_rate) {
        wasm.beatmapattributesbuilder_set_clock_rate(this.__wbg_ptr, !isLikeNone(clock_rate), isLikeNone(clock_rate) ? 0 : clock_rate);
    }
    /**
     * @param {boolean | null} [is_convert]
     */
    set isConvert(is_convert) {
        wasm.beatmapattributesbuilder_set_is_convert(this.__wbg_ptr, isLikeNone(is_convert) ? 0xFFFFFF : is_convert ? 1 : 0);
    }
    /**
     * Create a new `BeatmapAttributesBuilder`.
     * @param {BeatmapAttributesArgs | null} [args]
     */
    constructor(args) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.beatmapattributesbuilder_new(retptr, isLikeNone(args) ? 0 : addHeapObject(args));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            var r2 = getDataViewMemory0().getInt32(retptr + 4 * 2, true);
            if (r2) {
                throw takeObject(r1);
            }
            this.__wbg_ptr = r0 >>> 0;
            BeatmapAttributesBuilderFinalization.register(this, this.__wbg_ptr, this);
            return this;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Calculate the `BeatmapAttributes`.
     * @returns {BeatmapAttributes}
     */
    build() {
        const ptr = this.__destroy_into_raw();
        const ret = wasm.beatmapattributesbuilder_build(ptr);
        return BeatmapAttributes.__wrap(ret);
    }
    /**
     * @param {number | null} [ar]
     */
    set ar(ar) {
        wasm.beatmapattributesbuilder_set_ar(this.__wbg_ptr, isLikeNone(ar) ? 0x100000001 : Math.fround(ar));
    }
    /**
     * @param {number | null} [cs]
     */
    set cs(cs) {
        wasm.beatmapattributesbuilder_set_cs(this.__wbg_ptr, isLikeNone(cs) ? 0x100000001 : Math.fround(cs));
    }
    /**
     * @param {number | null} [hp]
     */
    set hp(hp) {
        wasm.beatmapattributesbuilder_set_hp(this.__wbg_ptr, isLikeNone(hp) ? 0x100000001 : Math.fround(hp));
    }
    /**
     * @param {number | null} [od]
     */
    set od(od) {
        wasm.beatmapattributesbuilder_set_od(this.__wbg_ptr, isLikeNone(od) ? 0x100000001 : Math.fround(od));
    }
    /**
     * @param {Beatmap | null} [map]
     */
    set map(map) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.beatmapattributesbuilder_set_map(retptr, this.__wbg_ptr, isLikeNone(map) ? 0 : addHeapObject(map));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            if (r1) {
                throw takeObject(r0);
            }
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @param {GameMode | null} [mode]
     */
    set mode(mode) {
        wasm.beatmapattributesbuilder_set_mode(this.__wbg_ptr, isLikeNone(mode) ? 4 : mode);
    }
    /**
     * @param {Object | null} [mods]
     */
    set mods(mods) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.beatmapattributesbuilder_set_mods(retptr, this.__wbg_ptr, isLikeNone(mods) ? 0 : addHeapObject(mods));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            if (r1) {
                throw takeObject(r0);
            }
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
}

const DifficultyFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_difficulty_free(ptr >>> 0, 1));
/**
 * Builder for a difficulty calculation.
 */
export class Difficulty {

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        DifficultyFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_difficulty_free(ptr, 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedAr(fixed) {
        wasm.difficulty_set_fixed_ar(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedCs(fixed) {
        wasm.difficulty_set_fixed_cs(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedHp(fixed) {
        wasm.difficulty_set_fixed_hp(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedOd(fixed) {
        wasm.difficulty_set_fixed_od(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {number | null} [clock_rate]
     */
    set clockRate(clock_rate) {
        wasm.difficulty_set_clock_rate(this.__wbg_ptr, !isLikeNone(clock_rate), isLikeNone(clock_rate) ? 0 : clock_rate);
    }
    /**
     * Returns a gradual difficulty calculator for the current difficulty settings.
     * @param {Beatmap} map
     * @returns {GradualDifficulty}
     */
    gradualDifficulty(map) {
        _assertClass(map, Beatmap);
        const ret = wasm.difficulty_gradualDifficulty(this.__wbg_ptr, map.__wbg_ptr);
        return GradualDifficulty.__wrap(ret);
    }
    /**
     * @param {number | null} [passed_objects]
     */
    set passedObjects(passed_objects) {
        wasm.difficulty_set_passed_objects(this.__wbg_ptr, isLikeNone(passed_objects) ? 0x100000001 : (passed_objects) >>> 0);
    }
    /**
     * Returns a gradual performance calculator for the current difficulty settings.
     * @param {Beatmap} map
     * @returns {GradualPerformance}
     */
    gradualPerformance(map) {
        _assertClass(map, Beatmap);
        const ret = wasm.difficulty_gradualPerformance(this.__wbg_ptr, map.__wbg_ptr);
        return GradualPerformance.__wrap(ret);
    }
    /**
     * @param {boolean | null} [hardrock_offsets]
     */
    set hardrockOffsets(hardrock_offsets) {
        wasm.difficulty_set_hardrock_offsets(this.__wbg_ptr, isLikeNone(hardrock_offsets) ? 0xFFFFFF : hardrock_offsets ? 1 : 0);
    }
    /**
     * Create a new difficulty calculator.
     * @param {DifficultyArgs | null} [args]
     */
    constructor(args) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.difficulty_new(retptr, isLikeNone(args) ? 0 : addHeapObject(args));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            var r2 = getDataViewMemory0().getInt32(retptr + 4 * 2, true);
            if (r2) {
                throw takeObject(r1);
            }
            this.__wbg_ptr = r0 >>> 0;
            DifficultyFinalization.register(this, this.__wbg_ptr, this);
            return this;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @param {number | null} [ar]
     */
    set ar(ar) {
        wasm.difficulty_set_ar(this.__wbg_ptr, isLikeNone(ar) ? 0x100000001 : Math.fround(ar));
    }
    /**
     * @param {number | null} [cs]
     */
    set cs(cs) {
        wasm.difficulty_set_cs(this.__wbg_ptr, isLikeNone(cs) ? 0x100000001 : Math.fround(cs));
    }
    /**
     * @param {number | null} [hp]
     */
    set hp(hp) {
        wasm.difficulty_set_hp(this.__wbg_ptr, isLikeNone(hp) ? 0x100000001 : Math.fround(hp));
    }
    /**
     * @param {number | null} [od]
     */
    set od(od) {
        wasm.difficulty_set_od(this.__wbg_ptr, isLikeNone(od) ? 0x100000001 : Math.fround(od));
    }
    /**
     * Perform the difficulty calculation but instead of evaluating strain
     * values, return them as is.
     *
     * Suitable to plot the difficulty over time.
     * @param {Beatmap} map
     * @returns {Strains}
     */
    strains(map) {
        _assertClass(map, Beatmap);
        const ret = wasm.difficulty_strains(this.__wbg_ptr, map.__wbg_ptr);
        return Strains.__wrap(ret);
    }
    /**
     * @param {Object | null} [mods]
     */
    set mods(mods) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.difficulty_set_mods(retptr, this.__wbg_ptr, isLikeNone(mods) ? 0 : addHeapObject(mods));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            if (r1) {
                throw takeObject(r0);
            }
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Perform the difficulty calculation.
     * @param {Beatmap} map
     * @returns {DifficultyAttributes}
     */
    calculate(map) {
        _assertClass(map, Beatmap);
        const ret = wasm.difficulty_calculate(this.__wbg_ptr, map.__wbg_ptr);
        return DifficultyAttributes.__wrap(ret);
    }
    /**
     * @param {boolean | null} [lazer]
     */
    set lazer(lazer) {
        wasm.difficulty_set_lazer(this.__wbg_ptr, isLikeNone(lazer) ? 0xFFFFFF : lazer ? 1 : 0);
    }
}

const DifficultyAttributesFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_difficultyattributes_free(ptr >>> 0, 1));
/**
 * The result of a difficulty calculation.
 */
export class DifficultyAttributes {

    static __wrap(ptr) {
        ptr = ptr >>> 0;
        const obj = Object.create(DifficultyAttributes.prototype);
        obj.__wbg_ptr = ptr;
        DifficultyAttributesFinalization.register(obj, obj.__wbg_ptr, obj);
        return obj;
    }

    toJSON() {
        return {
            mode: this.mode,
            stars: this.stars,
            isConvert: this.isConvert,
            aim: this.aim,
            aimDifficultSliderCount: this.aimDifficultSliderCount,
            speed: this.speed,
            flashlight: this.flashlight,
            sliderFactor: this.sliderFactor,
            aimTopWeightedSliderFactor: this.aimTopWeightedSliderFactor,
            speedTopWeightedSliderFactor: this.speedTopWeightedSliderFactor,
            speedNoteCount: this.speedNoteCount,
            aimDifficultStrainCount: this.aimDifficultStrainCount,
            speedDifficultStrainCount: this.speedDifficultStrainCount,
            nestedScorePerObject: this.nestedScorePerObject,
            legacyScoreBaseMultiplier: this.legacyScoreBaseMultiplier,
            maximumLegacyComboScore: this.maximumLegacyComboScore,
            hp: this.hp,
            nCircles: this.nCircles,
            nSliders: this.nSliders,
            nLargeTicks: this.nLargeTicks,
            nSpinners: this.nSpinners,
            stamina: this.stamina,
            rhythm: this.rhythm,
            color: this.color,
            reading: this.reading,
            nFruits: this.nFruits,
            nDroplets: this.nDroplets,
            nTinyDroplets: this.nTinyDroplets,
            nObjects: this.nObjects,
            nHoldNotes: this.nHoldNotes,
            ar: this.ar,
            preempt: this.preempt,
            greatHitWindow: this.greatHitWindow,
            okHitWindow: this.okHitWindow,
            mehHitWindow: this.mehHitWindow,
            monoStaminaFactor: this.monoStaminaFactor,
            mechanicalDifficulty: this.mechanicalDifficulty,
            consistencyFactor: this.consistencyFactor,
            maxCombo: this.maxCombo,
        };
    }

    toString() {
        return JSON.stringify(this);
    }

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        DifficultyAttributesFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_difficultyattributes_free(ptr, 0);
    }
    /**
     * The attributes' gamemode.
     * @returns {GameMode}
     */
    get mode() {
        const ret = wasm.__wbg_get_difficultyattributes_mode(this.__wbg_ptr);
        return ret;
    }
    /**
     * The final star rating.
     * @returns {number}
     */
    get stars() {
        const ret = wasm.__wbg_get_difficultyattributes_stars(this.__wbg_ptr);
        return ret;
    }
    /**
     * Whether the map was a convert i.e. an osu! map.
     * @returns {boolean}
     */
    get isConvert() {
        const ret = wasm.__wbg_get_difficultyattributes_isConvert(this.__wbg_ptr);
        return ret !== 0;
    }
    /**
     * The difficulty of the aim skill.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get aim() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_arHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The number of sliders weighted by difficulty.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get aimDifficultSliderCount() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odPerfectHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The difficulty of the speed skill.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get speed() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odGreatHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The difficulty of the flashlight skill.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get flashlight() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odGoodHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The ratio of the aim strain with and without considering sliders
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get sliderFactor() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odOkHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Describes how much of aim's difficult strain count is contributed to by sliders
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get aimTopWeightedSliderFactor() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_beatmapattributes_odMehHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Describes how much of speed's difficult strain count is contributed to by sliders
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get speedTopWeightedSliderFactor() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_speedTopWeightedSliderFactor(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The number of clickable objects weighted by difficulty.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get speedNoteCount() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_speedNoteCount(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Weighted sum of aim strains.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get aimDifficultStrainCount() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_aimDifficultStrainCount(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Weighted sum of speed strains.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get speedDifficultStrainCount() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_speedDifficultStrainCount(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The amount of nested score per object.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get nestedScorePerObject() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_nestedScorePerObject(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The legacy score base multiplier.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get legacyScoreBaseMultiplier() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_legacyScoreBaseMultiplier(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The maximum legacy combo score.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get maximumLegacyComboScore() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_maximumLegacyComboScore(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The health drain rate.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get hp() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_hp(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The amount of circles.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get nCircles() {
        const ret = wasm.__wbg_get_difficultyattributes_nCircles(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The amount of sliders.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get nSliders() {
        const ret = wasm.__wbg_get_difficultyattributes_nSliders(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The amount of "large ticks".
     *
     * The meaning depends on the kind of score:
     * - if set on osu!stable, this value is irrelevant
     * - if set on osu!lazer *with* slider accuracy, this value is the amount
     *   of hit slider ticks and repeats
     * - if set on osu!lazer *without* slider accuracy, this value is the
     *   amount of hit slider heads, ticks, and repeats
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get nLargeTicks() {
        const ret = wasm.__wbg_get_difficultyattributes_nLargeTicks(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The amount of spinners.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get nSpinners() {
        const ret = wasm.__wbg_get_difficultyattributes_nSpinners(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The difficulty of the stamina skill.
     *
     * Only available for osu!taiko.
     * @returns {number | undefined}
     */
    get stamina() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_stamina(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The difficulty of the rhythm skill.
     *
     * Only available for osu!taiko.
     * @returns {number | undefined}
     */
    get rhythm() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_rhythm(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The difficulty of the color skill.
     *
     * Only available for osu!taiko.
     * @returns {number | undefined}
     */
    get color() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_color(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The difficulty of the reading skill.
     *
     * Only available for osu!taiko.
     * @returns {number | undefined}
     */
    get reading() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_reading(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The amount of fruits.
     *
     * Only available for osu!catch.
     * @returns {number | undefined}
     */
    get nFruits() {
        const ret = wasm.__wbg_get_difficultyattributes_nFruits(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The amount of droplets.
     *
     * Only available for osu!catch.
     * @returns {number | undefined}
     */
    get nDroplets() {
        const ret = wasm.__wbg_get_difficultyattributes_nDroplets(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The amount of tiny droplets.
     *
     * Only available for osu!catch.
     * @returns {number | undefined}
     */
    get nTinyDroplets() {
        const ret = wasm.__wbg_get_difficultyattributes_nTinyDroplets(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The amount of hitobjects in the map.
     *
     * Only available for osu!mania.
     * @returns {number | undefined}
     */
    get nObjects() {
        const ret = wasm.__wbg_get_difficultyattributes_nObjects(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The amount of hold notes in the map.
     *
     * Only available for osu!mania.
     * @returns {number | undefined}
     */
    get nHoldNotes() {
        const ret = wasm.__wbg_get_difficultyattributes_nHoldNotes(this.__wbg_ptr);
        return ret === 0x100000001 ? undefined : ret;
    }
    /**
     * The approach rate.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get ar() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_ar(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Time preempt (AR time window).
     *
     * Only available for osu!catch.
     * @returns {number | undefined}
     */
    get preempt() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_preempt(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The perceived hit window for an n300 inclusive of rate-adjusting mods
     * (DT/HT/etc)
     *
     * Only available for osu! and osu!taiko.
     * @returns {number | undefined}
     */
    get greatHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_greatHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The perceived hit window for an n100 inclusive of rate-adjusting mods
     * (DT/HT/etc)
     *
     * Only available for osu! and osu!taiko.
     * @returns {number | undefined}
     */
    get okHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_okHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The perceived hit window for an n50 inclusive of rate-adjusting mods
     * (DT/HT/etc)
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get mehHitWindow() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_mehHitWindow(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The ratio of stamina difficulty from mono-color (single color) streams to total
     * stamina difficulty.
     *
     * Only available for osu!taiko.
     * @returns {number | undefined}
     */
    get monoStaminaFactor() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_monoStaminaFactor(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The difficulty corresponding to the mechanical skills.
     *
     * This includes colour and stamina combined.
     *
     * Only available for osu!taiko.
     * @returns {number | undefined}
     */
    get mechanicalDifficulty() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_mechanicalDifficulty(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The factor corresponding to the consistency of a map.
     *
     * Only available for osu!taiko.
     * @returns {number | undefined}
     */
    get consistencyFactor() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_difficultyattributes_consistencyFactor(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Return the maximum combo.
     * @returns {number}
     */
    get maxCombo() {
        const ret = wasm.__wbg_get_difficultyattributes_maxCombo(this.__wbg_ptr);
        return ret >>> 0;
    }
}

const GradualDifficultyFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_gradualdifficulty_free(ptr >>> 0, 1));
/**
 * Gradually calculate difficulty attributes after each hitobject.
 */
export class GradualDifficulty {

    static __wrap(ptr) {
        ptr = ptr >>> 0;
        const obj = Object.create(GradualDifficulty.prototype);
        obj.__wbg_ptr = ptr;
        GradualDifficultyFinalization.register(obj, obj.__wbg_ptr, obj);
        return obj;
    }

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        GradualDifficultyFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_gradualdifficulty_free(ptr, 0);
    }
    /**
     * Returns the amount of remaining items.
     * @returns {number}
     */
    get nRemaining() {
        const ret = wasm.gradualdifficulty_nRemaining(this.__wbg_ptr);
        return ret >>> 0;
    }
    /**
     * @param {Difficulty} difficulty
     * @param {Beatmap} map
     */
    constructor(difficulty, map) {
        _assertClass(difficulty, Difficulty);
        _assertClass(map, Beatmap);
        const ret = wasm.difficulty_gradualDifficulty(difficulty.__wbg_ptr, map.__wbg_ptr);
        this.__wbg_ptr = ret >>> 0;
        GradualDifficultyFinalization.register(this, this.__wbg_ptr, this);
        return this;
    }
    /**
     * Returns the `n`th attributes of the iterator.
     *
     * Note that the count starts from zero, so `nth(0)` returns the first
     * value, `nth(1)` the second, and so on.
     * @param {number} n
     * @returns {DifficultyAttributes | undefined}
     */
    nth(n) {
        const ret = wasm.gradualdifficulty_nth(this.__wbg_ptr, n);
        return ret === 0 ? undefined : DifficultyAttributes.__wrap(ret);
    }
    /**
     * Advances the iterator and returns the next attributes.
     * @returns {DifficultyAttributes | undefined}
     */
    next() {
        const ret = wasm.gradualdifficulty_next(this.__wbg_ptr);
        return ret === 0 ? undefined : DifficultyAttributes.__wrap(ret);
    }
    /**
     * Advances the iterator to the end to collect all remaining attributes
     * into a list and return them.
     * @returns {DifficultyAttributes[]}
     */
    collect() {
        try {
            const ptr = this.__destroy_into_raw();
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.gradualdifficulty_collect(retptr, ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            var v1 = getArrayJsValueFromWasm0(r0, r1).slice();
            wasm.__wbindgen_export_3(r0, r1 * 4, 4);
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
}

const GradualPerformanceFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_gradualperformance_free(ptr >>> 0, 1));
/**
 * Gradually calculate performance attributes after each hitresult.
 */
export class GradualPerformance {

    static __wrap(ptr) {
        ptr = ptr >>> 0;
        const obj = Object.create(GradualPerformance.prototype);
        obj.__wbg_ptr = ptr;
        GradualPerformanceFinalization.register(obj, obj.__wbg_ptr, obj);
        return obj;
    }

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        GradualPerformanceFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_gradualperformance_free(ptr, 0);
    }
    /**
     * Returns the amount of remaining items.
     * @returns {number}
     */
    get nRemaining() {
        const ret = wasm.gradualperformance_nRemaining(this.__wbg_ptr);
        return ret >>> 0;
    }
    /**
     * @param {Difficulty} difficulty
     * @param {Beatmap} map
     */
    constructor(difficulty, map) {
        _assertClass(difficulty, Difficulty);
        _assertClass(map, Beatmap);
        const ret = wasm.difficulty_gradualPerformance(difficulty.__wbg_ptr, map.__wbg_ptr);
        this.__wbg_ptr = ret >>> 0;
        GradualPerformanceFinalization.register(this, this.__wbg_ptr, this);
        return this;
    }
    /**
     * Process everything up to the next `n`th hitobject and calculate the
     * performance attributes for the resulting score state.
     *
     * Note that the count is zero-indexed, so `n=0` will process 1 object,
     * `n=1` will process 2, and so on.
     * @param {ScoreState} state
     * @param {number} n
     * @returns {PerformanceAttributes | undefined}
     */
    nth(state, n) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.gradualperformance_nth(retptr, this.__wbg_ptr, addBorrowedObject(state), n);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            var r2 = getDataViewMemory0().getInt32(retptr + 4 * 2, true);
            if (r2) {
                throw takeObject(r1);
            }
            return r0 === 0 ? undefined : PerformanceAttributes.__wrap(r0);
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
            heap[stack_pointer++] = undefined;
        }
    }
    /**
     * Process the next hit object and calculate the performance attributes
     * for the resulting score state.
     * @param {ScoreState} state
     * @returns {PerformanceAttributes | undefined}
     */
    next(state) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.gradualperformance_next(retptr, this.__wbg_ptr, addBorrowedObject(state));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            var r2 = getDataViewMemory0().getInt32(retptr + 4 * 2, true);
            if (r2) {
                throw takeObject(r1);
            }
            return r0 === 0 ? undefined : PerformanceAttributes.__wrap(r0);
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
            heap[stack_pointer++] = undefined;
        }
    }
}

const PerformanceFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_performance_free(ptr >>> 0, 1));
/**
 * Builder for a performance calculation.
 */
export class Performance {

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        PerformanceFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_performance_free(ptr, 0);
    }
    /**
     * @param {number | null} [misses]
     */
    set misses(misses) {
        wasm.performance_set_misses(this.__wbg_ptr, isLikeNone(misses) ? 0x100000001 : (misses) >>> 0);
    }
    /**
     * @param {number | null} [n_geki]
     */
    set nGeki(n_geki) {
        wasm.performance_set_n_geki(this.__wbg_ptr, isLikeNone(n_geki) ? 0x100000001 : (n_geki) >>> 0);
    }
    /**
     * @param {number | null} [n_katu]
     */
    set nKatu(n_katu) {
        wasm.performance_set_n_katu(this.__wbg_ptr, isLikeNone(n_katu) ? 0x100000001 : (n_katu) >>> 0);
    }
    /**
     * @param {number | null} [accuracy]
     */
    set accuracy(accuracy) {
        wasm.performance_set_accuracy(this.__wbg_ptr, !isLikeNone(accuracy), isLikeNone(accuracy) ? 0 : accuracy);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedAr(fixed) {
        wasm.performance_set_fixed_ar(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedCs(fixed) {
        wasm.performance_set_fixed_cs(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedHp(fixed) {
        wasm.performance_set_fixed_hp(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {boolean | null} [fixed]
     */
    set fixedOd(fixed) {
        wasm.performance_set_fixed_od(this.__wbg_ptr, isLikeNone(fixed) ? 0xFFFFFF : fixed ? 1 : 0);
    }
    /**
     * @param {number | null} [clock_rate]
     */
    set clockRate(clock_rate) {
        wasm.performance_set_clock_rate(this.__wbg_ptr, !isLikeNone(clock_rate), isLikeNone(clock_rate) ? 0 : clock_rate);
    }
    /**
     * @param {number | null} [passed_objects]
     */
    set passedObjects(passed_objects) {
        wasm.performance_set_passed_objects(this.__wbg_ptr, isLikeNone(passed_objects) ? 0x100000001 : (passed_objects) >>> 0);
    }
    /**
     * @param {number | null} [slider_end_hits]
     */
    set sliderEndHits(slider_end_hits) {
        wasm.performance_set_slider_ends_hit(this.__wbg_ptr, isLikeNone(slider_end_hits) ? 0x100000001 : (slider_end_hits) >>> 0);
    }
    /**
     * @param {boolean | null} [hardrock_offsets]
     */
    set hardrockOffsets(hardrock_offsets) {
        wasm.performance_set_hardrock_offsets(this.__wbg_ptr, isLikeNone(hardrock_offsets) ? 0xFFFFFF : hardrock_offsets ? 1 : 0);
    }
    /**
     * @param {number | null} [large_tick_hits]
     */
    set largeTickHits(large_tick_hits) {
        wasm.performance_set_large_ticks_hits(this.__wbg_ptr, isLikeNone(large_tick_hits) ? 0x100000001 : (large_tick_hits) >>> 0);
    }
    /**
     * @param {number | null} [small_tick_hits]
     */
    set smallTickHits(small_tick_hits) {
        wasm.performance_set_small_ticks_hits(this.__wbg_ptr, isLikeNone(small_tick_hits) ? 0x100000001 : (small_tick_hits) >>> 0);
    }
    /**
     * @param {HitResultPriority | null} [hitresult_priority]
     */
    set hitresultPriority(hitresult_priority) {
        wasm.performance_set_hitresult_priority(this.__wbg_ptr, isLikeNone(hitresult_priority) ? 2 : hitresult_priority);
    }
    /**
     * @param {number | null} [legacy_total_score]
     */
    set legacyTotalScore(legacy_total_score) {
        wasm.performance_set_legacy_total_score(this.__wbg_ptr, isLikeNone(legacy_total_score) ? 0x100000001 : (legacy_total_score) >>> 0);
    }
    /**
     * @param {HitResultGenerator | null} [hitresult_generator]
     * @param {GameMode | null} [mode]
     */
    setHitresultGenerator(hitresult_generator, mode) {
        wasm.performance_setHitresultGenerator(this.__wbg_ptr, isLikeNone(hitresult_generator) ? 2 : hitresult_generator, isLikeNone(mode) ? 4 : mode);
    }
    /**
     * Create a new performance calculator.
     * @param {PerformanceArgs | null} [args]
     */
    constructor(args) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.performance_new(retptr, isLikeNone(args) ? 0 : addHeapObject(args));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            var r2 = getDataViewMemory0().getInt32(retptr + 4 * 2, true);
            if (r2) {
                throw takeObject(r1);
            }
            this.__wbg_ptr = r0 >>> 0;
            PerformanceFinalization.register(this, this.__wbg_ptr, this);
            return this;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @param {number | null} [ar]
     */
    set ar(ar) {
        wasm.performance_set_ar(this.__wbg_ptr, isLikeNone(ar) ? 0x100000001 : Math.fround(ar));
    }
    /**
     * @param {number | null} [cs]
     */
    set cs(cs) {
        wasm.performance_set_cs(this.__wbg_ptr, isLikeNone(cs) ? 0x100000001 : Math.fround(cs));
    }
    /**
     * @param {number | null} [hp]
     */
    set hp(hp) {
        wasm.performance_set_hp(this.__wbg_ptr, isLikeNone(hp) ? 0x100000001 : Math.fround(hp));
    }
    /**
     * @param {number | null} [od]
     */
    set od(od) {
        wasm.performance_set_od(this.__wbg_ptr, isLikeNone(od) ? 0x100000001 : Math.fround(od));
    }
    /**
     * @param {number | null} [n50]
     */
    set n50(n50) {
        wasm.performance_set_n50(this.__wbg_ptr, isLikeNone(n50) ? 0x100000001 : (n50) >>> 0);
    }
    /**
     * @param {Object | null} [mods]
     */
    set mods(mods) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.performance_set_mods(retptr, this.__wbg_ptr, isLikeNone(mods) ? 0 : addHeapObject(mods));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            if (r1) {
                throw takeObject(r0);
            }
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @param {number | null} [n100]
     */
    set n100(n100) {
        wasm.performance_set_n100(this.__wbg_ptr, isLikeNone(n100) ? 0x100000001 : (n100) >>> 0);
    }
    /**
     * @param {number | null} [n300]
     */
    set n300(n300) {
        wasm.performance_set_n300(this.__wbg_ptr, isLikeNone(n300) ? 0x100000001 : (n300) >>> 0);
    }
    /**
     * Calculate performance attributes.
     *
     * If a beatmap is passed as argument, difficulty attributes will have to
     * be calculated internally which is a comparably expensive task. Hence,
     * passing previously calculated attributes should be prefered whenever
     * available.
     *
     * However, be careful that the passed attributes have been calculated
     * for the same difficulty settings like mods, clock rate, beatmap,
     * custom ar, ... otherwise the final attributes will be incorrect.
     * @param {MapOrAttributes} args
     * @returns {PerformanceAttributes}
     */
    calculate(args) {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.performance_calculate(retptr, this.__wbg_ptr, addBorrowedObject(args));
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            var r2 = getDataViewMemory0().getInt32(retptr + 4 * 2, true);
            if (r2) {
                throw takeObject(r1);
            }
            return PerformanceAttributes.__wrap(r0);
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
            heap[stack_pointer++] = undefined;
        }
    }
    /**
     * @param {number | null} [combo]
     */
    set combo(combo) {
        wasm.performance_set_combo(this.__wbg_ptr, isLikeNone(combo) ? 0x100000001 : (combo) >>> 0);
    }
    /**
     * @param {boolean | null} [lazer]
     */
    set lazer(lazer) {
        wasm.performance_set_lazer(this.__wbg_ptr, isLikeNone(lazer) ? 0xFFFFFF : lazer ? 1 : 0);
    }
}

const PerformanceAttributesFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_performanceattributes_free(ptr >>> 0, 1));
/**
 * The result of a performance calculation.
 */
export class PerformanceAttributes {

    static __wrap(ptr) {
        ptr = ptr >>> 0;
        const obj = Object.create(PerformanceAttributes.prototype);
        obj.__wbg_ptr = ptr;
        PerformanceAttributesFinalization.register(obj, obj.__wbg_ptr, obj);
        return obj;
    }

    toJSON() {
        return {
            difficulty: this.difficulty,
            state: this.state,
            pp: this.pp,
            ppAim: this.ppAim,
            ppFlashlight: this.ppFlashlight,
            ppSpeed: this.ppSpeed,
            ppAccuracy: this.ppAccuracy,
            effectiveMissCount: this.effectiveMissCount,
            estimatedUnstableRate: this.estimatedUnstableRate,
            speedDeviation: this.speedDeviation,
            comboBasedEstimatedMissCount: this.comboBasedEstimatedMissCount,
            scoreBasedEstimatedMissCount: this.scoreBasedEstimatedMissCount,
            aimEstimatedSliderBreaks: this.aimEstimatedSliderBreaks,
            speedEstimatedSliderBreaks: this.speedEstimatedSliderBreaks,
            ppDifficulty: this.ppDifficulty,
        };
    }

    toString() {
        return JSON.stringify(this);
    }

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        PerformanceAttributesFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_performanceattributes_free(ptr, 0);
    }
    /**
     * The difficulty attributes.
     * @returns {DifficultyAttributes}
     */
    get difficulty() {
        const ret = wasm.__wbg_get_performanceattributes_difficulty(this.__wbg_ptr);
        return DifficultyAttributes.__wrap(ret);
    }
    /**
     * The hitresult score state that was used for performance calculation.
     *
     * Only available if *not* created through gradual calculation.
     * @returns {ScoreState | undefined}
     */
    get state() {
        const ret = wasm.__wbg_get_performanceattributes_state(this.__wbg_ptr);
        return takeObject(ret);
    }
    /**
     * The final performance points.
     * @returns {number}
     */
    get pp() {
        const ret = wasm.__wbg_get_performanceattributes_pp(this.__wbg_ptr);
        return ret;
    }
    /**
     * The aim portion of the final pp.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get ppAim() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_ppAim(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The flashlight portion of the final pp.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get ppFlashlight() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_ppFlashlight(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The speed portion of the final pp.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get ppSpeed() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_ppSpeed(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The accuracy portion of the final pp.
     *
     * Only available for osu! and osu!taiko.
     * @returns {number | undefined}
     */
    get ppAccuracy() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_ppAccuracy(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Scaled miss count based on total hits.
     *
     * Only available for osu!.
     * @returns {number | undefined}
     */
    get effectiveMissCount() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_effectiveMissCount(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Upper bound on the player's tap deviation.
     *
     * Only *optionally* available for osu!taiko.
     * @returns {number | undefined}
     */
    get estimatedUnstableRate() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_estimatedUnstableRate(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Approximated unstable-rate
     *
     * Only *optionally* available for osu!.
     * @returns {number | undefined}
     */
    get speedDeviation() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_speedDeviation(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @returns {number | undefined}
     */
    get comboBasedEstimatedMissCount() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_comboBasedEstimatedMissCount(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @returns {number | undefined}
     */
    get scoreBasedEstimatedMissCount() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_scoreBasedEstimatedMissCount(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @returns {number | undefined}
     */
    get aimEstimatedSliderBreaks() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_aimEstimatedSliderBreaks(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * @returns {number | undefined}
     */
    get speedEstimatedSliderBreaks() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_speedEstimatedSliderBreaks(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * The strain portion of the final pp.
     *
     * Only available for osu!taiko and osu!mania.
     * @returns {number | undefined}
     */
    get ppDifficulty() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_performanceattributes_ppDifficulty(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r2 = getDataViewMemory0().getFloat64(retptr + 8 * 1, true);
            return r0 === 0 ? undefined : r2;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
}

const StrainsFinalization = (typeof FinalizationRegistry === 'undefined')
    ? { register: () => {}, unregister: () => {} }
    : new FinalizationRegistry(ptr => wasm.__wbg_strains_free(ptr >>> 0, 1));
/**
 * The result of calculating the strains of a beatmap.
 *
 * Suitable to plot the difficulty over time.
 */
export class Strains {

    static __wrap(ptr) {
        ptr = ptr >>> 0;
        const obj = Object.create(Strains.prototype);
        obj.__wbg_ptr = ptr;
        StrainsFinalization.register(obj, obj.__wbg_ptr, obj);
        return obj;
    }

    toJSON() {
        return {
            mode: this.mode,
            sectionLength: this.sectionLength,
            aim: this.aim,
            aimNoSliders: this.aimNoSliders,
            speed: this.speed,
            flashlight: this.flashlight,
            color: this.color,
            reading: this.reading,
            rhythm: this.rhythm,
            stamina: this.stamina,
            singleColorStamina: this.singleColorStamina,
            movement: this.movement,
            strains: this.strains,
        };
    }

    toString() {
        return JSON.stringify(this);
    }

    __destroy_into_raw() {
        const ptr = this.__wbg_ptr;
        this.__wbg_ptr = 0;
        StrainsFinalization.unregister(this);
        return ptr;
    }

    free() {
        const ptr = this.__destroy_into_raw();
        wasm.__wbg_strains_free(ptr, 0);
    }
    /**
     * The strains' gamemode.
     * @returns {GameMode}
     */
    get mode() {
        const ret = wasm.__wbg_get_strains_mode(this.__wbg_ptr);
        return ret;
    }
    /**
     * Time inbetween two strains in ms.
     * @returns {number}
     */
    get sectionLength() {
        const ret = wasm.__wbg_get_strains_sectionLength(this.__wbg_ptr);
        return ret;
    }
    /**
     * Strain peaks of the aim skill in osu!.
     * @returns {Float64Array | undefined}
     */
    get aim() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_aim(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the aim skill without sliders in osu!.
     * @returns {Float64Array | undefined}
     */
    get aimNoSliders() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_aimNoSliders(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the speed skill in osu!.
     * @returns {Float64Array | undefined}
     */
    get speed() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_speed(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the flashlight skill in osu!.
     * @returns {Float64Array | undefined}
     */
    get flashlight() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_flashlight(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the color skill in osu!taiko.
     * @returns {Float64Array | undefined}
     */
    get color() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_color(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the reading skill in osu!taiko.
     * @returns {Float64Array | undefined}
     */
    get reading() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_reading(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the rhythm skill in osu!taiko.
     * @returns {Float64Array | undefined}
     */
    get rhythm() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_rhythm(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the stamina skill in osu!taiko.
     * @returns {Float64Array | undefined}
     */
    get stamina() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_stamina(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the single color stamina skill in osu!taiko.
     * @returns {Float64Array | undefined}
     */
    get singleColorStamina() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_singleColorStamina(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the movement skill in osu!catch.
     * @returns {Float64Array | undefined}
     */
    get movement() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_movement(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
    /**
     * Strain peaks of the strain skill in osu!mania.
     * @returns {Float64Array | undefined}
     */
    get strains() {
        try {
            const retptr = wasm.__wbindgen_add_to_stack_pointer(-16);
            wasm.__wbg_get_strains_strains(retptr, this.__wbg_ptr);
            var r0 = getDataViewMemory0().getInt32(retptr + 4 * 0, true);
            var r1 = getDataViewMemory0().getInt32(retptr + 4 * 1, true);
            let v1;
            if (r0 !== 0) {
                v1 = getArrayF64FromWasm0(r0, r1).slice();
                wasm.__wbindgen_export_3(r0, r1 * 8, 8);
            }
            return v1;
        } finally {
            wasm.__wbindgen_add_to_stack_pointer(16);
        }
    }
}

async function __wbg_load(module, imports) {
    if (typeof Response === 'function' && module instanceof Response) {
        if (typeof WebAssembly.instantiateStreaming === 'function') {
            try {
                return await WebAssembly.instantiateStreaming(module, imports);

            } catch (e) {
                if (module.headers.get('Content-Type') != 'application/wasm') {
                    console.warn("`WebAssembly.instantiateStreaming` failed because your server does not serve Wasm with `application/wasm` MIME type. Falling back to `WebAssembly.instantiate` which is slower. Original error:\n", e);

                } else {
                    throw e;
                }
            }
        }

        const bytes = await module.arrayBuffer();
        return await WebAssembly.instantiate(bytes, imports);

    } else {
        const instance = await WebAssembly.instantiate(module, imports);

        if (instance instanceof WebAssembly.Instance) {
            return { instance, module };

        } else {
            return instance;
        }
    }
}

function __wbg_get_imports() {
    const imports = {};
    imports.wbg = {};
    imports.wbg.__wbg_buffer_609cc3eee51ed158 = function(arg0) {
        const ret = getObject(arg0).buffer;
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_call_672a4d21634d4a24 = function() { return handleError(function (arg0, arg1) {
        const ret = getObject(arg0).call(getObject(arg1));
        return addHeapObject(ret);
    }, arguments) };
    imports.wbg.__wbg_difficultyattributes_new = function(arg0) {
        const ret = DifficultyAttributes.__wrap(arg0);
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_done_769e5ede4b31c67b = function(arg0) {
        const ret = getObject(arg0).done;
        return ret;
    };
    imports.wbg.__wbg_entries_3265d4158b33e5dc = function(arg0) {
        const ret = Object.entries(getObject(arg0));
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_get_67b2ba62fc30de12 = function() { return handleError(function (arg0, arg1) {
        const ret = Reflect.get(getObject(arg0), getObject(arg1));
        return addHeapObject(ret);
    }, arguments) };
    imports.wbg.__wbg_get_b9b93047fe3cf45b = function(arg0, arg1) {
        const ret = getObject(arg0)[arg1 >>> 0];
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_getwithrefkey_2ba87e556248ad3d = function(arg0, arg1) {
        const ret = getObject(arg0)[getObject(arg1)];
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_instanceof_Uint8Array_17156bcf118086a9 = function(arg0) {
        let result;
        try {
            result = getObject(arg0) instanceof Uint8Array;
        } catch (_) {
            result = false;
        }
        const ret = result;
        return ret;
    };
    imports.wbg.__wbg_isArray_a1eab7e0d067391b = function(arg0) {
        const ret = Array.isArray(getObject(arg0));
        return ret;
    };
    imports.wbg.__wbg_isSafeInteger_343e2beeeece1bb0 = function(arg0) {
        const ret = Number.isSafeInteger(getObject(arg0));
        return ret;
    };
    imports.wbg.__wbg_iterator_9a24c88df860dc65 = function() {
        const ret = Symbol.iterator;
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_length_a446193dc22c12f8 = function(arg0) {
        const ret = getObject(arg0).length;
        return ret;
    };
    imports.wbg.__wbg_length_e2d2a49132c1b256 = function(arg0) {
        const ret = getObject(arg0).length;
        return ret;
    };
    imports.wbg.__wbg_new_405e22f390576ce2 = function() {
        const ret = new Object();
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_new_a12002a7f91c75be = function(arg0) {
        const ret = new Uint8Array(getObject(arg0));
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_new_c68d7209be747379 = function(arg0, arg1) {
        const ret = new Error(getStringFromWasm0(arg0, arg1));
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_next_25feadfc0913fea9 = function(arg0) {
        const ret = getObject(arg0).next;
        return addHeapObject(ret);
    };
    imports.wbg.__wbg_next_6574e1a8a62d1055 = function() { return handleError(function (arg0) {
        const ret = getObject(arg0).next();
        return addHeapObject(ret);
    }, arguments) };
    imports.wbg.__wbg_set_4c6c65d3d6468e8a = function(arg0, arg1, arg2) {
        getObject(arg0)[takeObject(arg1)] = takeObject(arg2);
    };
    imports.wbg.__wbg_set_65595bdd868b3009 = function(arg0, arg1, arg2) {
        getObject(arg0).set(getObject(arg1), arg2 >>> 0);
    };
    imports.wbg.__wbg_value_cd1ffa7b1ab794f1 = function(arg0) {
        const ret = getObject(arg0).value;
        return addHeapObject(ret);
    };
    imports.wbg.__wbindgen_as_number = function(arg0) {
        const ret = +getObject(arg0);
        return ret;
    };
    imports.wbg.__wbindgen_boolean_get = function(arg0) {
        const v = getObject(arg0);
        const ret = typeof(v) === 'boolean' ? (v ? 1 : 0) : 2;
        return ret;
    };
    imports.wbg.__wbindgen_debug_string = function(arg0, arg1) {
        const ret = debugString(getObject(arg1));
        const ptr1 = passStringToWasm0(ret, wasm.__wbindgen_export_1, wasm.__wbindgen_export_2);
        const len1 = WASM_VECTOR_LEN;
        getDataViewMemory0().setInt32(arg0 + 4 * 1, len1, true);
        getDataViewMemory0().setInt32(arg0 + 4 * 0, ptr1, true);
    };
    imports.wbg.__wbindgen_in = function(arg0, arg1) {
        const ret = getObject(arg0) in getObject(arg1);
        return ret;
    };
    imports.wbg.__wbindgen_is_function = function(arg0) {
        const ret = typeof(getObject(arg0)) === 'function';
        return ret;
    };
    imports.wbg.__wbindgen_is_object = function(arg0) {
        const val = getObject(arg0);
        const ret = typeof(val) === 'object' && val !== null;
        return ret;
    };
    imports.wbg.__wbindgen_is_undefined = function(arg0) {
        const ret = getObject(arg0) === undefined;
        return ret;
    };
    imports.wbg.__wbindgen_jsval_loose_eq = function(arg0, arg1) {
        const ret = getObject(arg0) == getObject(arg1);
        return ret;
    };
    imports.wbg.__wbindgen_memory = function() {
        const ret = wasm.memory;
        return addHeapObject(ret);
    };
    imports.wbg.__wbindgen_number_get = function(arg0, arg1) {
        const obj = getObject(arg1);
        const ret = typeof(obj) === 'number' ? obj : undefined;
        getDataViewMemory0().setFloat64(arg0 + 8 * 1, isLikeNone(ret) ? 0 : ret, true);
        getDataViewMemory0().setInt32(arg0 + 4 * 0, !isLikeNone(ret), true);
    };
    imports.wbg.__wbindgen_number_new = function(arg0) {
        const ret = arg0;
        return addHeapObject(ret);
    };
    imports.wbg.__wbindgen_object_clone_ref = function(arg0) {
        const ret = getObject(arg0);
        return addHeapObject(ret);
    };
    imports.wbg.__wbindgen_object_drop_ref = function(arg0) {
        takeObject(arg0);
    };
    imports.wbg.__wbindgen_string_get = function(arg0, arg1) {
        const obj = getObject(arg1);
        const ret = typeof(obj) === 'string' ? obj : undefined;
        var ptr1 = isLikeNone(ret) ? 0 : passStringToWasm0(ret, wasm.__wbindgen_export_1, wasm.__wbindgen_export_2);
        var len1 = WASM_VECTOR_LEN;
        getDataViewMemory0().setInt32(arg0 + 4 * 1, len1, true);
        getDataViewMemory0().setInt32(arg0 + 4 * 0, ptr1, true);
    };
    imports.wbg.__wbindgen_string_new = function(arg0, arg1) {
        const ret = getStringFromWasm0(arg0, arg1);
        return addHeapObject(ret);
    };
    imports.wbg.__wbindgen_throw = function(arg0, arg1) {
        throw new Error(getStringFromWasm0(arg0, arg1));
    };

    return imports;
}

function __wbg_init_memory(imports, memory) {

}

function __wbg_finalize_init(instance, module) {
    wasm = instance.exports;
    __wbg_init.__wbindgen_wasm_module = module;
    cachedDataViewMemory0 = null;
    cachedFloat64ArrayMemory0 = null;
    cachedUint8ArrayMemory0 = null;



    return wasm;
}

function initSync(module) {
    if (wasm !== undefined) return wasm;


    if (typeof module !== 'undefined') {
        if (Object.getPrototypeOf(module) === Object.prototype) {
            ({module} = module)
        } else {
            console.warn('using deprecated parameters for `initSync()`; pass a single object instead')
        }
    }

    const imports = __wbg_get_imports();

    __wbg_init_memory(imports);

    if (!(module instanceof WebAssembly.Module)) {
        module = new WebAssembly.Module(module);
    }

    const instance = new WebAssembly.Instance(module, imports);

    return __wbg_finalize_init(instance, module);
}

async function __wbg_init(module_or_path) {
    if (wasm !== undefined) return wasm;


    if (typeof module_or_path !== 'undefined') {
        if (Object.getPrototypeOf(module_or_path) === Object.prototype) {
            ({module_or_path} = module_or_path)
        } else {
            console.warn('using deprecated parameters for the initialization function; pass a single object instead')
        }
    }

    if (typeof module_or_path === 'undefined') {
        module_or_path = new URL('rosu_pp_js_bg.wasm', import.meta.url);
    }
    const imports = __wbg_get_imports();

    if (typeof module_or_path === 'string' || (typeof Request === 'function' && module_or_path instanceof Request) || (typeof URL === 'function' && module_or_path instanceof URL)) {
        module_or_path = fetch(module_or_path);
    }

    __wbg_init_memory(imports);

    const { instance, module } = await __wbg_load(await module_or_path, imports);

    return __wbg_finalize_init(instance, module);
}

export { initSync };
export default __wbg_init;
