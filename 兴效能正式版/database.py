import os
import pandas as pd
import glob
from datetime import datetime
import time
import globals

class ExcelDataUpdater:
    def __init__(self, target_folder="data_source", keyword="兴效能"):
        """
        初始化数据更新器
        
        参数:
            target_folder: 监控的子目录名称 (默认: "data_source")
            keyword: 文件名关键词 (默认: "兴效能")
        """
        self.current_dir = os.path.dirname(os.path.abspath(__file__))
        self.data_source_dir = os.path.join(self.current_dir, target_folder)
        self.keyword = keyword
        self.last_update_time = None
        self.cached_data = {}
        
        # 确保目录存在
        if not os.path.exists(self.data_source_dir):
            os.makedirs(self.data_source_dir)
            print(f"创建了目录: {self.data_source_dir}")

    def get_latest_excel_file(self):
        """获取最新的符合条件的Excel文件"""
        excel_files = glob.glob(os.path.join(self.data_source_dir, f'*{self.keyword}*.xlsx')) + \
                      glob.glob(os.path.join(self.data_source_dir, f'*{self.keyword}*.xls'))
        
        if not excel_files:
            return None
        
        # 获取最新的文件
        latest_file = max(excel_files, key=os.path.getmtime)
        return latest_file

    def read_all_sheets(self, file_path):
        """读取Excel文件中所有sheet页的数据（从第二行开始）"""
        try:
            excel_file = pd.ExcelFile(file_path)
            sheets_data = {}
            
            for sheet_name in excel_file.sheet_names:
                df = pd.read_excel(
                    excel_file,
                    sheet_name=sheet_name,
                    header=1  # 从第二行开始读取
                )
                
                if not df.empty:
                    sheets_data[sheet_name] = df
                    print(f"  → 已加载: {sheet_name} ({len(df)}行)")
                else:
                    print(f"  → 警告: {sheet_name} 无数据")
            
            return sheets_data
        
        except Exception as e:
            print(f"读取Excel文件失败: {e}")
            return None

    def check_and_update(self):
        """检查并更新数据"""
        latest_file = self.get_latest_excel_file()
        
        if not latest_file:
            print("未找到符合条件的Excel文件")
            return False
        
        file_mod_time = os.path.getmtime(latest_file)
        
        # 如果是第一次运行或文件有更新
        if self.last_update_time is None or file_mod_time > self.last_update_time:
            print(f"\n检测到文件更新: {os.path.basename(latest_file)}")
            print(f"最后修改时间: {datetime.fromtimestamp(file_mod_time)}")
            
            # 读取新数据
            new_data = self.read_all_sheets(latest_file)
            
            if new_data:
                self.cached_data = new_data
                self.last_update_time = file_mod_time
                print("数据更新完成!")
                return True
            else:
                print("数据更新失败")
                return False
        else:
            print("没有检测到新文件更新")
            return False

    def get_data(self, sheet_name=None):
        """获取缓存的数据"""
        if sheet_name:
            return self.cached_data.get(sheet_name)
        return self.cached_data

    def start_monitoring(self, interval=60):
        """启动定时监控"""
        print(f"\n开始监控目录: {self.data_source_dir}")
        print(f"监控关键词: '{self.keyword}'")
        print(f"检查间隔: {interval}秒")
        print("按Ctrl+C停止监控...")
        
        try:
            while True:
                self.check_and_update()
                time.sleep(interval)
        except KeyboardInterrupt:
            print("\n监控已停止")

# 使用示例
if __name__ == "__main__":
    # 创建更新器实例
    updater = ExcelDataUpdater(target_folder="data_source", keyword="兴效能")
    
    # 初始更新
    updater.check_and_update()
    
    # 获取特定sheet数据示例
    flight_data = updater.get_data("进离港航班")
    if flight_data is not None:
        print("\n进离港航班数据预览:")
        print(flight_data.head())
    
    # 启动定时监控（可选）
    updater.start_monitoring(interval=60)  # 每1分钟检查一次