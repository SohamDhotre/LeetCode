/**
 * Definition for singly-linked list.
 * public class ListNode {
 *     int val;
 *     ListNode next;
 *     ListNode() {}
 *     ListNode(int val) { this.val = val; }
 *     ListNode(int val, ListNode next) { this.val = val; this.next = next; }
 * }
 */
class Solution {
    public ListNode[] splitListToParts(ListNode head, int k) {
        // if(head==null || head.next==null || k==1) return new ListNode[]{head};
        ListNode []ans=new ListNode[k];
        int len=len(head);
        int firstMaxLists=len%k;
        int minElements=len/k;
        int itr=0;
        ListNode prev=null, cur=head;
        while(itr<k){
            int count=0, allowedElements;
            // store the head of each sub list
            ans[itr]=cur;
            allowedElements=(itr<firstMaxLists)?minElements+1:minElements;
            while(count<allowedElements){
                // allow only N/K+1 elements for starting N%K lists
                prev=cur;
                cur=cur.next;
                count++;
            }
            if(prev!=null) prev.next=null;
            itr++;
        }
        return ans;
    }

    int len(ListNode head){
        ListNode cur=head;
        int len=0;
        while(cur!=null){
            cur=cur.next;
            len++;
        }
        return len;
    }
}