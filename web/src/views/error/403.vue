<template>
  <div class="error-page">
    <div class="error-content">
      <h1>403</h1>
      <p>抱歉，您没有权限访问此页面</p>
      <el-button type="primary" @click="$router.push('/')">返回首页</el-button>
    </div>
  </div>
</template>

<style scoped lang="scss">
.error-page {
  display: flex;
  justify-content: center;
  align-items: center;
  height: 100%;
}

.error-content {
  text-align: center;

  h1 {
    font-size: 72px;
    color: #e6a23c;
    margin-bottom: 16px;
  }

  p {
    font-size: 16px;
    color: #909399;
    margin-bottom: 24px;
  }
}
</style>
