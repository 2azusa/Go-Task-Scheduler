package handler

import (
	"crony/admin/internal/model/request"
	"crony/admin/internal/model/resp"
	"crony/admin/internal/service"
	"crony/common/models"
	"crony/common/pkg/etcdclient"
	"crony/common/pkg/logger"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/gin-gonic/gin"
)

// NodeRouter 结构体用于定义相关的路由处理器
type NodeRouter struct{}

// defaultNodeRouter 是 NodeRouter 的一个全局单例，在注册路由时使用
var defaultNodeRouter = new(NodeRouter)

// Delete 方法用于处理删除节点的HTTP请求
func (n *NodeRouter) Delete(c *gin.Context) {
	// 1. 解析请求参数
	var req request.ByUUID // 定义一个变量，用于接收包含UUID的请求体
	// 将请求的JOSN体绑定到req变量上
	if err := c.ShouldBindJSON(&req); err != nil {
		// 记录错误日志
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] request parameter error: %s", err.Error()))
		// 返回参数错误响应
		resp.FailWithMessage(resp.ErrorRequestParameter, "[delete_node] request parameter error", c)
	}

	// 2. 根据UUID从数据库中查找节点信息
	node := &models.Node{UUID: req.UUID}
	err := node.FindByUUID()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] find node by uuid: %s error: %s", req.UUID, err.Error()))
		resp.FailWithMessage(resp.ERROR, "[deleta_node db find error]", c)
		return
	}

	// 3. 逻辑校验，不能删除一个处于存活的节点
	if node.Status == models.NodeConnSuccess {
		resp.FailWithMessage(resp.ERROR, "[delete_node] can't delete a node that is already alive", c)
		return
	}

	// 4. 清理etcd中的相关数据，删除该节点下的所有任务定义
	// 使用 `WithPrefix()` 可以删除 /crony/job/profile/<node_uuid>/ 前缀下的所有键值对
	_, _ = etcdclient.Delete(fmt.Sprintf(etcdclient.KeyEtcdJobProfile, req.UUID), clientv3.WithPrefix())

	// 5. 从数据库中删除节点记录
	err = node.Delete()
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[delete_node] into db error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[delete_node] db delete error", c)
		return
	}

	// 6. 返回成功响应
	resp.OkWithMessage("delete success", c)
}

// Serach 方法用于处理搜索节点的HTTP请求
func (n *NodeRouter) Search(c *gin.Context) {
	// 1. 节点和校验请求参数
	var req request.ReqNodeSearch // 定义变量，用于接收节点的搜索请求
	// 绑定JSON请求体到req变量
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_node] request parameter error: %s", err.Error()))
		resp.FailWithMessage(resp.ErrorRequestParameter, "[search_node] request parameter error", c)
		return
	}
	// 调用 Check() 方法，为分页等参数设置默认值，避免空指针
	req.Check()

	// 2. 调用服务层执行搜索逻辑
	// 将搜索条件传递给服务层的 Search 方法，获取节点列表、总数和可能发生的错误
	nodes, total, err := service.DefaultNodeWatcher.Search(&req)
	if err != nil {
		logger.GetLogger().Error(fmt.Sprintf("[search_node] search node error: %s", err.Error()))
		resp.FailWithMessage(resp.ERROR, "[search_node] search node error", c)
		return
	}

	// 3. 组装响应数据
	var resultNodes []resp.RspNodeSearch // 定义一个用于响应的节点列表切片
	// 遍历从数据库查出的节点列表
	for _, node := range nodes {
		// 为每个节点创建一个响应专用的结构体
		resultNode := resp.RspNodeSearch{
			Node: node, // 将数据库模型赋值给响应模型的Node字段
		}
		// 调用服务层方法，获取该节点当前承载的任务数量
		resultNode.JobCount, _ = service.DefaultNodeWatcher.GetJobCount(node.UUID)
		// 将组装好的节点信息追加到响应列表中
		resultNodes = append(resultNodes, resultNode)
	}

	// 4. 返回包含分页信息的成功响应
	resp.OkWithDetailed(resp.PageResult{
		List:     resultNodes,  // 最终的节点列表
		Total:    total,        // 满足条件的总记录数
		Page:     req.Page,     // 当前页码
		PageSize: req.PageSize, // 每页大小
	}, "search success", c)
}
