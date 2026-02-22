import re

with open('/Users/zhangjun/CursorProjects/Macda-Connector/web-nb67-web/src/views/airConditioner/trainUnitMapData.vue', 'r') as f:
    content = f.read()

# Check if it works
print("Found bg div:", '<div class="bg"><img src="/src/assets/img/TrainUnitMapBg.svg" /></div>' in content)
