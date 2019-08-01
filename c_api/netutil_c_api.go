package main
/*
#include <stdint.h>
#include <stdbool.h>

// Mapped from app-netutil.lib/v1alpha/types.go

struct CPUResponse {
    char*    CPUSet;
};


struct EnvData {
    char*    Index;
    char*    Value;
};
// *pEnvs is an array of 'struct EnvData' that is allocated
// from the C program.
struct EnvResponse {
	int             netutil_num_envs;
	struct EnvData *pEnvs;
};

#define NETUTIL_NUM_IPS			10
#define NETUTIL_NUM_NETWORKSTATUS	10
#define NETUTIL_NUM_NETWORKINTERFACE	10

struct NetworkStatus {
    char*    Name;
    char*    Interface;
    char*    IPs[NETUTIL_NUM_IPS];
    char*    Mac;
};
struct NetworkStatusResponse {
	struct NetworkStatus Status[NETUTIL_NUM_NETWORKSTATUS];
};

struct SriovData {
	char*	PCIAddress;
};

struct VhostData {
	char*	SocketFile;
	bool	Master;
};

struct NetworkInterface {
	char*	Name;
	char*	Type;
	struct	SriovData	Sriov;
	struct	VhostData	Vhost;
};
struct NetworkInterfaceResponse {
	struct NetworkInterface	Interface[NETUTIL_NUM_NETWORKINTERFACE];
};

*/
import "C"
import "unsafe"

import (
	"flag"

	"github.com/golang/glog"

	netlib "github.com/openshift/app-netutil/lib/v1alpha"
)

const (
	cpusetPath = "/sys/fs/cgroup/cpuset/cpuset.cpus"
	netutil_num_ips = 10
	netutil_num_networkstatus = 10
	netutil_num_networkinterface = 10

	// Interface type
	NETUTIL_INTERFACE_TYPE_PCI = "pci"
	NETUTIL_INTERFACE_TYPE_VHOST = "vhost"

	// Errno
	NETUTIL_ERRNO_SUCCESS = 0
	NETUTIL_ERRNO_FAIL = 1
	NETUTIL_ERRNO_SIZE_ERROR = 2
)


//export GetCPUInfo
func GetCPUInfo(c_cpuResp *C.struct_CPUResponse) int64 {
	flag.Parse()
	cpuRsp, err := netlib.GetCPUInfo()

	if err == nil {
		c_cpuResp.CPUSet = C.CString(cpuRsp.CPUSet)
		return NETUTIL_ERRNO_SUCCESS
	}
	glog.Errorf("netlib.GetCPUInfo() err: %+v", err)
	return NETUTIL_ERRNO_FAIL
}

//export GetEnv
func GetEnv(c_envResp *C.struct_EnvResponse) int64 {
	var j _Ctype_int

	flag.Parse()
	envRsp, err := netlib.GetEnv()

	if err == nil {
		j = 0

		// Map the input pointer to array of structures, c_envResp.pEnvs, to
		// a slice of the structures, c_envResp_pEnvs. Then the slice can be
		// indexed.
		c_envResp_pEnvs := (*[1 << 30]C.struct_EnvData)(unsafe.Pointer(c_envResp.pEnvs))[:c_envResp.netutil_num_envs:c_envResp.netutil_num_envs]
		for i, env := range envRsp.Envs {
			if j < c_envResp.netutil_num_envs {
				c_envResp_pEnvs[j].Index = C.CString(i)
				c_envResp_pEnvs[j].Value = C.CString(env)
				j++
			} else {
				glog.Errorf("EnvResponse struct not sized properly. At %d ENV Variables.", j)
				return NETUTIL_ERRNO_SIZE_ERROR
			}
		}
		return NETUTIL_ERRNO_SUCCESS
	}
	glog.Errorf("netlib.GetEnv() err: %+v", err)
	return NETUTIL_ERRNO_FAIL
}


//export GetNetworkStatus
func GetNetworkStatus(c_networkResp *C.struct_NetworkStatusResponse) int64 {
	flag.Parse()
	networkStatusRsp, err := netlib.GetNetworkStatus()

	if err == nil {
		for i, networkStatus := range networkStatusRsp.Status {
			if i < netutil_num_networkstatus {
				c_networkResp.Status[i].Name = C.CString(networkStatus.Name)
				c_networkResp.Status[i].Interface = C.CString(networkStatus.Interface)
				c_networkResp.Status[i].Mac = C.CString(networkStatus.Mac)
				for j, ipaddr := range networkStatus.IPs {
					if j < netutil_num_ips {
						c_networkResp.Status[i].IPs[j] = C.CString(ipaddr)
					} else {
						glog.Errorf("NetworkStatusResponse IPs struct" +
							"not sized properly. At %d IPs for Interface %d.", j, i)
						return NETUTIL_ERRNO_SIZE_ERROR
					}
				}
			} else {
				glog.Errorf("NetworkStatusResponse struct not sized properly." +
					"At %d Interfaces.", i)
				return NETUTIL_ERRNO_SIZE_ERROR
			}
		}
		return NETUTIL_ERRNO_SUCCESS
	}
	glog.Errorf("netlib.GetNetworkStatus() err: %+v", err)
	return NETUTIL_ERRNO_FAIL
}

//export GetNetworkInterface
func GetNetworkInterface(c_intType *C.char, c_netIntResp *C.struct_NetworkInterfaceResponse) int64 {
	flag.Parse()
	intRsp, err := netlib.GetNetworkInterface(C.GoString(c_intType))

	if err == nil {
		for i, iface := range intRsp.Interface {
			if i < netutil_num_networkinterface {
				c_netIntResp.Interface[i].Name = C.CString(iface.Name)
				c_netIntResp.Interface[i].Type = C.CString(iface.Type)
				switch C.GoString(c_intType) {

				case NETUTIL_INTERFACE_TYPE_PCI:
					c_netIntResp.Interface[i].Sriov.PCIAddress =
						C.CString(iface.Sriov.PCIAddress)

				case NETUTIL_INTERFACE_TYPE_VHOST:
					c_netIntResp.Interface[i].Vhost.SocketFile =
						C.CString(iface.Vhost.SocketFile)
					c_netIntResp.Interface[i].Vhost.Master =
						(iface.Vhost.Master != false)

				case "":
					c_netIntResp.Interface[i].Sriov.PCIAddress =
						C.CString(iface.Sriov.PCIAddress)
					c_netIntResp.Interface[i].Vhost.SocketFile =
						C.CString(iface.Vhost.SocketFile)
					c_netIntResp.Interface[i].Vhost.Master =
						(iface.Vhost.Master != false)
				}
			} else {
				glog.Errorf("NetworkInterfaceResponse struct not sized properly." +
					"At %d Interfaces.", i)
				return NETUTIL_ERRNO_SIZE_ERROR
			}
		}
		return NETUTIL_ERRNO_SUCCESS
	}
	glog.Errorf("netlib.GetNetworkInterface() err: %+v", err)
	return NETUTIL_ERRNO_FAIL
}

func main() {}
