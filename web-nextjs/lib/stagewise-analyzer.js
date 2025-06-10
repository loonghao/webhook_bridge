/**
 * Webpack plugin to analyze stagewise bundle impact
 */

class StagewiseAnalyzerPlugin {
  constructor(options = {}) {
    this.options = {
      enabled: options.enabled !== false,
      outputFile: options.outputFile || 'stagewise-analysis.json',
      ...options
    };
  }

  apply(compiler) {
    if (!this.options.enabled) {
      return;
    }

    compiler.hooks.emit.tapAsync('StagewiseAnalyzerPlugin', (compilation, callback) => {
      const stats = compilation.getStats().toJson();
      const analysis = this.analyzeBundle(stats);
      
      // Output analysis
      const analysisJson = JSON.stringify(analysis, null, 2);
      compilation.assets[this.options.outputFile] = {
        source: () => analysisJson,
        size: () => analysisJson.length
      };

      // Log summary
      this.logSummary(analysis);
      
      callback();
    });
  }

  analyzeBundle(stats) {
    const stagewiseModules = [];
    const totalSize = stats.assets.reduce((sum, asset) => sum + asset.size, 0);
    let stagewiseSize = 0;

    // Find stagewise-related modules
    if (stats.modules) {
      stats.modules.forEach(module => {
        if (this.isStagewiseModule(module)) {
          stagewiseModules.push({
            name: module.name,
            size: module.size,
            chunks: module.chunks
          });
          stagewiseSize += module.size || 0;
        }
      });
    }

    // Find stagewise-related assets
    const stagewiseAssets = stats.assets.filter(asset => 
      this.isStagewiseAsset(asset.name)
    );

    return {
      timestamp: new Date().toISOString(),
      environment: process.env.NODE_ENV || 'development',
      stagewiseEnabled: process.env.NEXT_PUBLIC_ENABLE_STAGEWISE === 'true',
      totalBundleSize: totalSize,
      stagewiseSize: stagewiseSize,
      stagewisePercentage: totalSize > 0 ? (stagewiseSize / totalSize * 100).toFixed(2) : 0,
      stagewiseModules: stagewiseModules,
      stagewiseAssets: stagewiseAssets.map(asset => ({
        name: asset.name,
        size: asset.size
      })),
      recommendations: this.generateRecommendations(stagewiseSize, totalSize)
    };
  }

  isStagewiseModule(module) {
    const name = module.name || '';
    return name.includes('stagewise') || 
           name.includes('StagewiseProvider') ||
           name.includes('StagewiseDebugger') ||
           name.includes('useStagewise');
  }

  isStagewiseAsset(assetName) {
    return assetName.includes('stagewise') || 
           assetName.includes('debug');
  }

  generateRecommendations(stagewiseSize, totalSize) {
    const recommendations = [];
    const percentage = totalSize > 0 ? (stagewiseSize / totalSize * 100) : 0;

    if (percentage > 5) {
      recommendations.push({
        type: 'warning',
        message: `Stagewise占用了 ${percentage.toFixed(2)}% 的打包体积，考虑在生产环境中禁用`
      });
    }

    if (process.env.NODE_ENV === 'production' && process.env.NEXT_PUBLIC_ENABLE_STAGEWISE === 'true') {
      recommendations.push({
        type: 'warning',
        message: '生产环境中启用了 Stagewise，这可能影响性能'
      });
    }

    if (stagewiseSize === 0 && process.env.NODE_ENV === 'production') {
      recommendations.push({
        type: 'success',
        message: 'Stagewise 已成功从生产构建中移除'
      });
    }

    return recommendations;
  }

  logSummary(analysis) {
    console.log('\n📊 Stagewise Bundle Analysis');
    console.log('================================');
    console.log(`Environment: ${analysis.environment}`);
    console.log(`Stagewise Enabled: ${analysis.stagewiseEnabled}`);
    console.log(`Total Bundle Size: ${(analysis.totalBundleSize / 1024).toFixed(2)} KB`);
    console.log(`Stagewise Size: ${(analysis.stagewiseSize / 1024).toFixed(2)} KB`);
    console.log(`Stagewise Percentage: ${analysis.stagewisePercentage}%`);
    
    if (analysis.recommendations.length > 0) {
      console.log('\n💡 Recommendations:');
      analysis.recommendations.forEach(rec => {
        const icon = rec.type === 'warning' ? '⚠️' : '✅';
        console.log(`${icon} ${rec.message}`);
      });
    }
    
    console.log('================================\n');
  }
}

module.exports = StagewiseAnalyzerPlugin;
