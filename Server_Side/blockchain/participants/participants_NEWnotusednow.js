/*eslint-env node */
var create = require(__dirname+'/CRUD/create.js');
exports.create = create.create;

var read = require(__dirname+'/CRUD/read.js');
exports.read = read.read;

var regulatorsFile = require(__dirname+'/regulators/regulators.js');
var regulators = {};
regulators.read = regulatorsFile.read;
exports.regulators = regulators;

var suppliersFile = require(__dirname+'/suppliers/suppliers.js');
var suppliers = {};
suppliers.read = suppliersFile.read;
exports.suppliers = suppliers;

var warehouseFile = require(__dirname+'/warehouse/warehouse.js');
var warehouse = {};
warehouse.read = warehouseFile.read;
exports.warehouse = warehouse;

var airportsFile = require(__dirname+'/airports/airports.js');
var airports = {};
airports.read = airportsFile.read;
exports.airports = airports;

var deliveryFile = require(__dirname+'/delivery/delivery.js');
var delivery = {};
delivery.read = deliveryFile.read;
exports.delivery = delivery;

var buyerFile = require(__dirname+'/buyer/buyer.js');
var buyer = {};
buyer.read = buyerFile.read;
exports.buyer = buyer;