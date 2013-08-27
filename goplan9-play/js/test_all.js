test("asString", function(){

	var call = new exports.Fcall();
	call.Type = exports.MessageType.Tversion;
	call.Version = "9p2000.u";
	call.Tag = exports.NOTAG;
	call.Msize = 0;

	ok("Tversion tag 65535 msize 0 version '9p2000.u'" === call.asString(), "Tversion asString");
	call.Type = exports.MessageType.Rversion;
	ok("Rversion tag 65535 msize 0 version '9p2000.u'" === call.asString(), "Rversion asString");

});

test("qid asString", function(){
	var qid = new exports.Qid();
	qid.Vers = 0;
	qid.Type = exports.QTDIR;
	qid.Path = exports.Uint64.fromInteger(1); // usually root
	ok("(l1:h0 0 d)" === qid.asString(), "qid asString");
});
