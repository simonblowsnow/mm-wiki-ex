
var ExcelCompress = {
    compress: compress,
    decompress: decompress
}

function decompress (dts) {
    dts.forEach(item => {
        var d = item.data;
        if (Array.isArray(d) || d.size == undefined) return;
        var data = [];
        for (var i = 0; i < d.size[0]; i++) {
            var line = [], flag = IsEmpty(d.c[i]);
            for (var j = 0; j < d.size[1]; j++) {
                var v = flag ? null : (IsEmpty(d.c[i][j]) ? null : d.c[i][j]);
                line.push(v);
            }
            data.push(line);
        }
        item.data = data;
    });
    
}
function IsEmpty(e) {
    return e == null || e == undefined;
}

function compress (dts) {
    dts.forEach(d => {
        if (!Array.isArray(d.data) || d.data.length == 0 || !Array.isArray(d.data[0])) return;
        d.data = {'size': [d.data.length, d.data[0].length], 'c': compressData(d.data)};
    });
} 

function compressData (data) {
    var res = {};
    data.forEach((d, i) => {
        var line = {};
        d.forEach((e, j) => {
            if (e != null) line[j] = e;
        });
        if (Object.keys(line).length > 0) res[i] = line;
    });
    return res;
}


// 针对复杂情况的压缩，多个默认值
function compressDataEx (data) {
    var r = _checkItem(data, true);
    var res, c;
    r.next(); // JS协程的机制，需要先进入
    // process rows
    data.forEach((d, i) => {
        // process cols
        c = _checkItem(d, false).next().value;
        if (c['nulls'].length == 1 && c['nulls'][0][0] == 0 && c['nulls'][0][1] == (d.length - 1)) c = null;
        res = r.next(c);
    });
    res = r.next(c);
    return res.value;
}

// 若大量值为空值，或只有一种默认值，不需记录连续值，不记录入字典即可
function* _checkItem(d, wait) {
    var c = {"nulls": []}, ns = [];
    var n = d.length - 1, empty = null;
    for (var j = 0; j < d.length; j++) {
        var e = wait ? yield : d[j];
        if (e == empty) {
            if (ns.length == 0) ns.push(j);
        } else {
            c[j] = e;
            if (ns.length == 1) ns = _set(c, ns, j - 1);
        } 
    }
    var last = wait ? yield : d[n];
    if (ns.length == 1 && last == empty) _set(c, ns, n);
    return c;
}

function _set(c, ns, i) {
    if (ns[0] == i) {
        c[i] = null;
    } else {
        c['nulls'].push([ns[0], i]);
    }
    return [];  // 只是为了少写一行代码
}

