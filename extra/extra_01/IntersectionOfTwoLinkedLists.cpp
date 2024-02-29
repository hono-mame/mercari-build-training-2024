struct ListNode {
    int val;
     ListNode *next;
    ListNode(int x) : val(x), next(nullptr) {}
};

class Solution {
public:
    int length(ListNode *head){
        int len = 0;
        while(head){
            len++;
            head = head -> next;
        }
        return len;
    }
    ListNode *getIntersectionNode(ListNode *headA, ListNode *headB) {
        if(headA == nullptr || headB == nullptr){
            return nullptr;
        }
        int lenA = length(headA);
        int lenB = length(headB);

        // 短い方の長さまでノードを進める
        if(lenA < lenB){
            while(lenA < lenB){
                headB = headB -> next;
                lenB--;
            }
        }
        if(lenB < lenA){
            while(lenB < lenA){
                headA = headA -> next;
                lenA--;
            }
        }

        // 共通祖先を見つけたらその時のノードを返す、違かったらノードを次に進める
        while(headA != nullptr && headB != nullptr){
            if(headA==headB) return headA;
            headA = headA->next;
            headB = headB->next;
            }
        return nullptr;
    }
};