<template>
  <div>
    <h2>工作台</h2>
    <el-row :gutter="20" style="margin-top: 20px;">
      <el-col :span="6">
        <el-card>
          <div style="text-align: center;">
            <div style="font-size: 28px; color: #409EFF;">0</div>
            <div style="color: #999; margin-top: 8px;">人员总数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div style="text-align: center;">
            <div style="font-size: 28px; color: #67C23A;">0</div>
            <div style="color: #999; margin-top: 8px;">在职人数</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div style="text-align: center;">
            <div style="font-size: 28px; color: #E6A23C;">0</div>
            <div style="color: #999; margin-top: 8px;">公司数量</div>
          </div>
        </el-card>
      </el-col>
      <el-col :span="6">
        <el-card>
          <div style="text-align: center;">
            <div style="font-size: 28px; color: #F56C6C;">0</div>
            <div style="color: #999; margin-top: 8px;">本月操作</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>
