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
    public ListNode swapNodes(ListNode head, int k) {
        if(head.next==null) return head;
        ListNode leftNode=head, rightNode=head, fast=head;
        for(int i=1;i<k;i++){
            leftNode=leftNode.next;
            fast=fast.next;
        }
        fast=fast.next;
        while(fast!=null){
            rightNode=rightNode.next;
            fast=fast.next;
        }
        
        int temp=leftNode.val;
        leftNode.val=rightNode.val;
        rightNode.val=temp;

        return head;
    }
}