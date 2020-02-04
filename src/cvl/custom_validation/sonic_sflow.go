package custom_validation

import (
	"net"
	util "cvl/internal/util"
 )


//Validate ip address of sflow collector ip 
func (t *CustomValidation) ValidateCollectorIp(vc *CustValidationCtxt) CVLErrorInfo {

	if (vc.CurCfg.VOp == OP_DELETE) {
		 return CVLErrorInfo{ErrCode: CVL_SUCCESS}
	}

	ip := net.ParseIP(vc.YNodeVal)
	if ip == nil {
		 errStr:= "Sflow collector IP is not valid"
	         util.CVL_LEVEL_LOG(util.ERROR,"%s",errStr)
                 return CVLErrorInfo{
			            ErrCode: CVL_SYNTAX_INVALID_INPUT_DATA,
				    TableName: "SFLOW_COLLECTOR",
				    CVLErrDetails : errStr,
				    ConstraintErrMsg : errStr,
			    }
	}

	if ip.IsLoopback() || ip.IsUnspecified() || ip.Equal(net.IPv4bcast) || ip.IsMulticast() {
		errStr:= "Sflow collector IP is not valid, IP is either reserved, unspecified, loopback, or broadcast"
		util.CVL_LEVEL_LOG(util.ERROR,"%s",errStr)
		return CVLErrorInfo{
			            ErrCode: CVL_SYNTAX_INVALID_INPUT_DATA,
				    TableName: "SFLOW_COLLECTOR",
				    CVLErrDetails : errStr,
				    ConstraintErrMsg : errStr,
			    }
	}
	return CVLErrorInfo{ErrCode: CVL_SUCCESS}
}

