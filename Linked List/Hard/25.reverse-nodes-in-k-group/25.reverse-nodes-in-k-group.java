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
    public ListNode reverseKGroup(ListNode head, int k) {
        if(head==null || head.next==null || k==1) return head;
        ListNode cur=head, prev=null;
        int count=0;
        while(count<k && cur!=null){
            prev=cur;
            cur=cur.next;
            count++;
        }
        if(count==k){
            ListNode newHead=reverseFirstK(head, k);
            head.next=reverseKGroup(cur, k);
            return newHead;
        }
        else{
            return head;
        }
    }

    ListNode reverseFirstK(ListNode head, int k){
        ListNode cur=head, prev=null;
        while(k>0){
            ListNode temp=cur.next;
            cur.next=prev;
            prev=cur;
            cur=temp;
            k--;
        }
        return prev;
    }
}