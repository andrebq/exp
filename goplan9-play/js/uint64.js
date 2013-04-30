
exports.Uint64 = function(low, high) {
	this.low = low;
	this.high = high;
};

exports.Uint64.prototype.asString = function() {
	return sprintf("l%u:h%u", this.low, this.high);
};

exports.Uint64.fromInteger = function(integer) {
	return new exports.Uint64(integer, 0);
};
