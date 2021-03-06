'use strict';

var base = '';
var priv = '6ERz8OH1V5o47ecPnBSbjqk3L2OkwZtT4OSDPfCj1Jg=';
var user = '6Ik+FbfTq+SgPisFIVH7fN3TDzzvtraHL8Oqc357ZEM=';
var users = {
    '6Ik+FbfTq+SgPisFIVH7fN3TDzzvtraHL8Oqc357ZEM=': '小小童鞋'
};
var curve = new elliptic.ec('p256');
var curTab;
var $navs = {};
var $tabs = {};
var channel;
var token;
var key;

function sendClick() {
    var value = $('#footbar-input').val();
    if (value !== '' && typeof curTab !== 'undefined') {
        send(token, key, value);
        $('#footbar-input').val('');
    }
}

function post(url, data, cb, fl) {
    $.ajax({
        type: 'POST', url: base + url, dataType: 'json',
        data: JSON.stringify(data), success: cb, error: fl
    });
}

function decodePoint(str) {
    var data = base64js.toByteArray(str);
    return curve.keyFromPublic(data).getPublic();
}

function decodePriv(str) {
    var data = base64js.toByteArray(str);
    return curve.keyFromPrivate(data).getPrivate();
}

function enter(user, cb) {
    post('/enter', { t: user }, function (data) {
        var c1 = decodePoint(data.c1);
        var c2 = decodePoint(data.c2);
        var p = decodePriv(priv);
        var m = c1.add(c2.mul(p).neg());
        var key = base64js.fromByteArray(m.encode());
        key = base64js.fromByteArray(sha256.array(key));
        if (typeof cb === 'function') cb(data.t, key);
    }, function () {
        loadDialogue();
    });
}

function encrypt(key, msg) {
    var iv = sjcl.random.randomWords(3);
    var data = sjcl.codec.utf8String.toBits(JSON.stringify(msg));
    key = sjcl.codec.base64.toBits(key);
    var aes = new sjcl.cipher.aes(key);
    var encrypted = sjcl.mode.gcm.encrypt(aes, data, iv);
    iv = sjcl.codec.base64.fromBits(iv);
    encrypted = sjcl.codec.base64.fromBits(encrypted);
    return { d: encrypted, i: iv };
}

function decrypt(key, msg) {
    var iv = sjcl.codec.base64.toBits(msg.i);
    var data = sjcl.codec.base64.toBits(msg.d);
    key = sjcl.codec.base64.toBits(key);
    var aes = new sjcl.cipher.aes(key);
    var decrypted = sjcl.mode.gcm.decrypt(aes, data, iv);
    decrypted = sjcl.codec.utf8String.fromBits(decrypted);
    return JSON.parse(decrypted);
}

function newtab(id, name) {
    var $main = $('<div></div>').addClass('main')
               .attr('id', 'main-' + id).hide().appendTo('#main');
    var $head = $('<div></div>').addClass('main-head').addClass('col')
               .appendTo($main);
    var $title = $('<a></a>').attr('href', '#')
                .addClass('main-title').text(name).appendTo($head);
    var $con = $('<div></div>').addClass('main-container')
              .addClass('col').addClass('scroll').appendTo($main);
    $title.click(function () {
        $('#dialogue-token').val(id);
        $('#dialogue-name').val($title.text());
        $('#dialogue-tab').show();
    });
    return { m: $main, c: $con, t: $title };
}

function newnav(id, name) {
    var $nav = $('<a></a>').attr('href', '#')
              .addClass('sidebar-link').text(name)
              .appendTo('#sidebar-list');
    $nav.click(function () {
        if (typeof curTab !== 'undefined') {
            $navs[curTab].removeClass('active');
            $tabs[curTab].m.hide();
        }
        curTab = id;
        $navs[curTab].addClass('active');
        $tabs[curTab].m.show();
    });
    return $nav;
}

function startnav() {
    $('#sidebar-list').empty();
    $('.main').remove();
    $('#sidebar-head').text(users[user] ? users[user] : '*');
    var $add = $('<a></a>').attr('href', '#').addClass('sidebar-link')
              .attr('id', 'sidebar-new').text('+')
              .appendTo($('#sidebar-list'));
    $add.click(function () {
        $('#dialogue-token-new').val('');
        $('#dialogue-name-new').val('');
        $('#dialogue-new').show();
    });
    for (var x in users) {
        var ch = users[x];
        $navs[x] = newnav(x, ch);
        $tabs[x] = newtab(x, ch);
    }
}

function appendMsg(channel, sender, msg) {
    var $con;
    if (channel === user) {
        $con = $tabs[sender].c;
    } else {
        $con = $tabs[channel].c;
    }
    var display = users[sender];
    if (typeof display === 'undefined') {
        display = '*';
    }
    if (typeof $con !== 'undefined') {
        var $msg = $('<div></div>').addClass('message').appendTo($con);
        var $sender = $('<p></p>').addClass('message-sender')
                 .text(display).appendTo($msg);
        var $data = $('<pre></pre>').addClass('message-data')
               .text(msg).appendTo($msg);
        $con.scrollTop($con.prop('scrollHeight'));
    }
}

function startenter() {
    token = '';
    enter(user, function(t, k) {
        token = t;
        key = k;
        poll(token, key);
    });
}

function poll(t, k) {
    if (token === t) {
        post('/poll', { t: t }, function (data) {
            for (var i = 0; i < data.length; i++) {
                var msg = decrypt(k, data[i]);
                var sender = msg.u;
                var channel = msg.c;
                var txt = msg.m;
                appendMsg(channel, sender, txt);
            }
            poll(t, k);
        }, function (_, msg) {
            if (token === t) {
                startenter();
            }
        });
    }
}

function send(token, key, txt) {
    var msg = { u: user, c: curTab, m: txt };
    var data = encrypt(key, msg);
    post('/send', { t: token, m: data });
}

function start() {
    var data = JSON.parse(localStorage.chaty);
    user = data.user;
    priv = data.priv;
    users = data.users;

    startnav();
    startenter();
}

function updatePub() {
    try {
        var priv = $('#dialogue-priv').val();
        priv = base64js.toByteArray(priv);
        var key = curve.keyFromPrivate(priv);
        var pub = key.getPublic().encode();
        pub = base64js.fromByteArray(pub);
        $('#dialogue-pub').val(pub);
    } catch(e) {
        $('#dialogue-pub').val('N/A');
    }
}

function clearDialogue() {
    $('#dialogue-priv').val('');
    $('#dialogue-user').val('');
    $('#dialogue-users').val('{}');
    updatePub();
}

function loadDialogue() {
    if (typeof localStorage.chaty !== 'undefined') {
        var data = JSON.parse(localStorage.chaty);
        $('#dialogue-priv').val(data.priv);
        $('#dialogue-user').val(data.user);
        $('#dialogue-users').val(JSON.stringify(data.users));
        updatePub();
    } else {
        clearDialogue();
    }
    $('#dialogue').show();
}

function saveUsers() {
    var data = JSON.parse(localStorage.chaty);
    data.users = users;
    localStorage.chaty = JSON.stringify(data);
}

$(document).ready(function () {
    $('#sidebar-head').click(loadDialogue);

    $('#dialogue-gen').click(function () {
        var user = sjcl.random.randomWords(8);
        user = sjcl.codec.base64.fromBits(user);
        var pair = curve.genKeyPair();
        var priv = pair.getPrivate().toArray();
        priv = base64js.fromByteArray(priv);
        $('#dialogue-priv').val(priv);
        $('#dialogue-user').val(user);
        updatePub();
    });

    $('#dialogue-cancel').click(function () {
        $('#dialogue').hide();
    });

    $('#dialogue-save').click(function () {
        var data = {
            priv:     $('#dialogue-priv').val(),
            user:     $('#dialogue-user').val(),
            users:    JSON.parse($('#dialogue-users').val())
        }
        localStorage.chaty = JSON.stringify(data);
        $('#dialogue').hide();

        start();
    });

    $('#dialogue-clear').click(clearDialogue);
    $('#dialogue-priv').blur(updatePub);
    $('#footbar-btn').click(sendClick);

    $('#dialogue-cancel-tab').click(function () {
        $('#dialogue-tab').hide();
    });

    $('#dialogue-save-tab').click(function () {
        var name = $('#dialogue-name').val();
        if (typeof curTab !== 'undefined') {
            users[curTab] = name;
            $tabs[curTab].t.text(name);
            $navs[curTab].text(name);
            if (user === curTab) {
                $('#sidebar-head').text(name);
            }
            saveUsers();
        }
        $('#dialogue-tab').hide();
    });

    $('#dialogue-del').click(function () {
        if (typeof curTab !== 'undefined') {
            delete users[curTab];
            $navs[curTab].remove();
            delete $navs[curTab];
            $tabs[curTab].m.remove();
            delete $tabs[curTab];
            if (user === curTab) {
                $('#sidebar-head').text('*');
            }
            curTab = undefined;
            saveUsers();
        }
        $('#dialogue-tab').hide();
    });

    $('#dialogue-cancel-new').click(function () {
        $('#dialogue-new').hide();
    });

    $('#dialogue-save-new').click(function () {
        var id = $('#dialogue-token-new').val();
        var name = $('#dialogue-name-new').val();
        if (typeof users[id] === 'undefined') {
            users[id] = name;
            $navs[id] = newnav(id, name);
            $tabs[id] = newtab(id, name);
            if (user === id) {
                $('#sidebar-head').text(name);
            }
            saveUsers();
        }
        $('#dialogue-new').hide();
    });

    $('#footbar-input').keydown(function (e) {
        if (e.which == 13) {
            if (e.ctrlKey || e.metaKey) {
                $('#footbar-input').val($('#footbar-input').val() + '\n');
            } else {
                sendClick();
            }
            e.preventDefault();
        }
    });

    if (typeof localStorage.chaty !== 'undefined') {
        start();
    } else {
        loadDialogue();
    }
});
