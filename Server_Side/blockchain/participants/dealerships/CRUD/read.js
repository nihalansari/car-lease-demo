/*eslint-env node */

var tracing = require(__dirname+'/../../../../tools/traces/trace.js');
var reload = require('require-reload')(require),
    participants = reload(__dirname+'/../../participants_info.js');

var read = function(req, res)
{
	participants = reload(__dirname+'/../../participants_info.js');
	tracing.create('ENTER', 'GET blockchain/participants/warehouse', {});
	
	if(!participants.participants_info.hasOwnProperty('warehouse'))
	{
		res.status(404)
		var error = {}
		error.message = 'Unable to retrieve warehouse'
		error.error = true;
		tracing.create('ERROR', 'GET blockchain/participants/warehouse', error);
		res.send(error)
	} 
	else
	{
		tracing.create('EXIT', 'GET blockchain/participants/warehouse', {"result":participants.participants_info.warehouse});
		res.send({"result":participants.participants_info.warehouse})
	}
}
exports.read = read;