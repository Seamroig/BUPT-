# BUPT-
区块链算法设计一种加密货币，参考自ivan kuznetsov博客
并未实现p2p网络，相比于ivan的源码也没有将矿工节点独立出来，只完成了一部分基本功能如pow算法，钱包地址，交易及其验证和网络同步功能等
采用了端口号nodeid代替了真实的ip地址

p.s：
Pow算法的Targetbit为16,适合cpu进行基本的test，一般十秒以内就可以得出结果
UI界面做成了中文，但是最后设置矿工节点的api部分功能不全
Final report中详细介绍了基于Bitcoin的Vcoin，并且提出了peer-to-peer网络设计——一种混合式的p2p网络，与此同时，Test的结果和部分截图也包含在内了



