
var remove = require(__dirname+'/CRUD/delete.js');
exports.delete = remove.delete;

var read = require(__dirname+'/CRUD/read.js');
exports.read = read.read;


var typeFile = require(__dirname+'/type/type.js');
var type = {};
type.update = typeFile.update;
type.read = typeFile.read;
exports.type = type;

var particularsFile = require(__dirname+'/particulars/particulars.js');
var particulars = {};
particulars.update = particularsFile.update;
particulars.read = particularsFile.read;
exports.particulars = particulars;

var sourcecityFile = require(__dirname+'/sourcecity/sourcecity.js');
var sourcecity = {};
sourcecity.update = sourcecityFile.update;
sourcecity.read = sourcecityFile.read;
exports.sourcecity = sourcecity;

var destcityFile = require(__dirname+'/destcity/destcity.js');
var destcity = {};
destcity.update = destcityFile.update;
destcity.read = destcityFile.read;
exports.destcity = destcity;

var weightFile = require(__dirname+'/weight/weight.js');
var weight = {};
weight.update = weightFile.update;
weight.read = weightFile.read;
exports.weight = weight;

var deliveredFile = require(__dirname+'/delivered/delivered.js');
var delivered = {};
delivered.update = deliveredFile.update;
delivered.read = deliveredFile.read;
exports.delivered = delivered;

var ownerFile = require(__dirname+'/owner/owner.js');
var owner = {};
owner.update = ownerFile.update;
owner.read = ownerFile.read;
exports.owner = owner;


var statFile = require(__dirname+'/status/status.js');
var stat = {};
stat.update = statFile.update;
stat.read = statFile.read;
exports.stat = stat;

var lastlocationFile = require(__dirname+'/lastlocation/lastlocation.js');
var lastlocation = {};
lastlocation.update = lastlocationFile.update;
lastlocation.read = lastlocationFile.read;
exports.lastlocation = lastlocation;

var dispatchdateFile = require(__dirname+'/dispatchdate/dispatchdate.js');
var dispatchdate = {};
dispatchdate.update = dispatchdateFile.update;
dispatchdate.read = dispatchdateFile.read;
exports.dispatchdate = dispatchdate;

var delivereddateFile = require(__dirname+'/delivereddate/delivereddate.js');
var delivereddate = {};
delivereddate.update = delivereddateFile.update;
delivereddate.read = delivereddateFile.read;
exports.delivereddate = delivereddate;

var dimensionsFile = require(__dirname+'/dimensions/dimensions.js');
var dimensions = {};
dimensions.update = dimensionsFile.update;
dimensions.read = dimensionsFile.read;
exports.dimensions = dimensions;

