登陆模块
（1）登陆接口
bool Login(int userid,string username,string passward)
（2）注册接口
bool Register(int userid,string username,string passward)
装备模块
（1）获取装备信息
string GetAllEquipFromMysql(int userid)
（2）保存装备信息
bool SaveAllEquipToMysql(int userid,string equipinfo)
（3）删除某一个装备
bool DelEquip(int userid,int equipid)
（4）删除某一品质的装备，用于删除对应玩家中某一品质的所有装备
bool BatchDelEquip(int userid,int quality)
角色信息模块
（1）获取角色信息
string GetPlayerInfo(int userid)
（2）更新玩家信息
bool UpdatePlayerinfo(int userid,string playerinfo)
好友模块
（1）获取所有好友信息
string GetAllFriendInfo(int userid)
（2）更新所有好友信息
string UpdateAllFriendInfo(int userid)
（3）加好友
bool AddFriend(int fromuserid,int touserid)
（4）删好友
bool DelFriend(int fromuserid,int touserid)
（5）好友申请：被申请方
bool FriendAskTo(int fromuserid,int touserid,bool isagree)
（6）好友申请：申请方：监听
bool FriendAskFrom(int fromuserid,int touserid,bool isagree)


排行榜模块
（1）获取所有排行榜信息
string GetAllRankInfo()
（2）获取个人排行信息
string GetSelfRankInfo(int userid)
数据库设计
（1）user
userid INT 主键 非空 唯一
username VARCHAR(45) 非空
password VARCHAR(45) 非空
（2）equip
equipid INT 主键 非空 唯一
quality INT 非空
damage INT
crit INT
critdamage INT
damagespeed INT
bllodsuck INT
hp INT
movespeed INT
equipname VARCHAR(45)
userid INT 非空
defense INT
goodfortune INT
（3）playerinfo
userid INT 主键 非空
level INT
experience INT
gamelevel INT
blood_energy INT

云端mysql链接如下：
string server = "rm-2zevr95ez9rrid70uho.mysql.rds.aliyuncs.com";
string database = "Vampire";
string user = "wwk18255113901";
string password = "BaiChen123456+";


