<template>
  <div style="text-align:center;padding:100px 0;">
    <el-result icon="error" title="404 / 403" sub-title="页面不存在或无权限访问">
      <template #extra>
        <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
      </template>
    </el-result>
  </div>
</template>
