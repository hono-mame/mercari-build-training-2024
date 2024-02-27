#include <vector>
using namespace std;
class Solution {
public:
    int eraseOverlapIntervals(vector<vector<int>>& intervals) {
        sort(intervals.begin(), intervals.end());
        int previous = 0;
        int answer = 0;
        for(int current=1;current<intervals.size();current++){
            
            // -> overlapしている時 [1,2],[1,3] など
            // これはどちらかが余分なのでans++する
            if(intervals[current][0]<intervals[previous][1]){
                answer++;

                // endの値の小さい方をpreviousに設定する(answerを最小化するため)
                if(intervals[current][1]<=intervals[previous][1]){
                    previous=current;
                }
            } 
            // overlapしていない時は確定させ、次を見る
            else {
                previous=current;
            }
        }
     return answer;
    }
};