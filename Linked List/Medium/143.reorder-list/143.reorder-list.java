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
    public void reorderList(ListNode head) {
        ListNode slow=head, fast=head;
        // find the mid node
        while(fast!=null && fast.next!=null){
            slow=slow.next;
            fast=fast.next.next;
        }
        // get the right part
        ListNode right=slow.next;
        slow.next=null;
        // rev the right list
        ListNode prev=null;
        ListNode cur=right;
        while(cur!=null){
            ListNode temp=cur.next;
            cur.next=prev;
            prev=cur;
            cur=temp;
        }
        // take one node from each list and link them alternatively
        ListNode revHead=prev;
        ListNode leftHead=head;
        while(revHead!=null){
            ListNode nextLeft=leftHead.next;
            ListNode nextRight=revHead.next;
            
            leftHead.next=revHead;
            revHead.next=nextLeft;

            leftHead=nextLeft;
            revHead=nextRight;
        }

        //return the same head
    }
}