#include <vector>
using namespace std;
class Solution {
public:
    int minEatingSpeed(vector<int>& piles, int h) {
        int left = 1;
        int right = *max_element(piles.begin(), piles.end()); // O(n)
        // 二分探索で最小のkを求める // O(logn)
        while (left < right) {
            int mid = (left + right) / 2;
            int time = 0;
            for (int pile : piles) { // O(n)
                time += (pile + mid - 1) / mid; // 切り上げ
            }
            if (time > h) {
                left = mid + 1; // スピードを上げる
            } else {
                right = mid; // スピードを下げる
            }
        }
        return left; // 最小速度を返す
    }
};
